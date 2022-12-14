// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2022 Uri Shaked <uri@wokwi.com>

package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/containers/gvisor-tap-vsock/pkg/virtualnetwork"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/sirupsen/logrus"
	"github.com/wokwi/wokwigw/pkg/loopback"

	"github.com/spf13/cobra"
)

var (
	version   = "unreleased"
	gitHash   = ""
	buildTime = ""

	config = defaultConfig()
)

var flags = flagCfg{
	listenPort:  defaultListenPort,
	forwardList: []string{},
}

func execute() error {
	cobra.MousetrapHelpText = ""
	return newRootCmd(&flags, &config).Execute()
}

func newRootCmd(flags *flagCfg, config *types.Configuration) *cobra.Command {

	rootCmd := &cobra.Command{
		Use:   "wokwigw",
		Short: fmt.Sprintf("wokwi IoT Gateway (ver:%s)", version),
		Long: fmt.Sprintf(`
wokwi IoT Gateway (ver:%s)

	Connect your Wokwi simulated IoT Devices (e.g. ESP32) to you local network!
`, version),
		RunE: run,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// You can bind cobra and viper in a few locations, but PersistencePreRunE on the root command works well
			return validateAndMapFlags(flags, config)
		},
	}

	// configure flags
	f := rootCmd.PersistentFlags()

	f.StringSliceVar(&flags.forwardList, "forward", flags.forwardList, "specify one or more forwarding localPort:address:remotePort tuples")
	f.IntVar(&flags.listenPort, "listenPort", flags.listenPort, "listening port (on local host)")

	return rootCmd
}

func validateAndMapFlags(flags *flagCfg, cfg *types.Configuration) (err error) {

	// map command line arguments to configuration structure (forwardList)
	// note: since we're using syntax similar to ssh -L option, we do the splitting ourselves here
	for _, fwd := range flags.forwardList {
		parts := strings.Split(fwd, ":")
		if len(parts) != 3 {
			return fmt.Errorf(string("arg %s is not formatted using the syntax 'localPort:addr:remotePort'"), fwd)
		}

		if v, e := strconv.Atoi(parts[0]); err != nil || v < 0 || v > 65535 {
			return fmt.Errorf("invalid local port specified in forward argument (%s): %w", fwd, e)
		}

		if v, e := strconv.Atoi(parts[2]); err != nil || v < 0 || v > 65535 {
			return fmt.Errorf("invalid remote port specified in forward argument (%s): %w", fwd, e)
		}

		cfg.Forwards[":"+parts[0]] = net.JoinHostPort(parts[1], parts[2])
	}

	if flags.listenPort < 0 || flags.listenPort > 65535 {
		return fmt.Errorf("invalid listen port specified (%d)", flags.listenPort)
	}

	return nil
}

func banner() {

	var gitStr string

	if gitHash != "" {
		gitStr = fmt.Sprintf("Git revision: %s\nBuilt: %s\n\n", gitHash, buildTime)
	}

	fmt.Printf(`
       __              ,
|  |  /  \  |_/  |  |  |
|/\|  \__/  | \  |/\|  |

    Wokwi IoT Gateway

Version: %s
%s
Listening on TCP Port %d
`, version, gitStr, flags.listenPort)
}

func run(cmd *cobra.Command, _ []string) error {
	banner()

	logrus.SetLevel(logrus.WarnLevel)

	vn, err := virtualnetwork.New(&config)
	if err != nil {
		return fmt.Errorf("error creating network %w", err)
	}

	err = http.ListenAndServe(net.JoinHostPort(defaultListenAddr, strconv.Itoa(flags.listenPort)), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		fmt.Printf("[%s] Client connected (%s)\n", r.RemoteAddr, origin)

		if !checkOrigin(origin) {
			w.WriteHeader(http.StatusForbidden)
			fmt.Printf("[%s] Invalid origin: %s\n", r.RemoteAddr, origin)
			return
		}

		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			fmt.Printf("[%s] Web socket error: %s\n", r.RemoteAddr, err)
			return
		}

		pipe1, pipe2, err := loopback.ConnLoopback()
		if err != nil {
			fmt.Printf("[%s] Pipe creation failed: %s\n", r.RemoteAddr, err)
			return
		}

		writer := wsutil.NewWriter(conn, ws.StateServerSide, ws.OpText)
		encoder := json.NewEncoder(writer)
		err = encoder.Encode(makeAlohaMessage(version))
		if err != nil {
			return
		}

		err = writer.Flush()
		if err != nil {
			fmt.Printf("[%s] Write error: %s\n", r.RemoteAddr, err)
			return
		}

		// need a wait group to wait for the below goroutines before exiting this function
		wg := sync.WaitGroup{}
		ctx, cancel := context.WithCancel(cmd.Context())

		wg.Add(2)
		cleanup := func() {
			cancel()

			// todo: need to handle errors here
			_ = conn.Close()
			_ = pipe1.Close()
			_ = pipe2.Close()
			wg.Done()
		}

		// todo: this function returns error - handle or wrap
		go vn.AcceptQemu(ctx, pipe1)

		go func() {
			defer cleanup()

			for {
				msg, op, err := wsutil.ReadClientData(conn)
				if err != nil {
					return
				}
				if op == ws.OpBinary {
					err := binary.Write(pipe2, binary.BigEndian, uint32(len(msg)))
					if err != nil {
						return
					}

					_, err = pipe2.Write(msg)
					if err != nil {
						return
					}
				} else if op == ws.OpText {
					fmt.Printf("[%s] Incoming message: %v\n", r.RemoteAddr, msg)
				}
			}
		}()

		go func() {
			defer cleanup()

			for {
				var size uint32
				err := binary.Read(pipe2, binary.BigEndian, &size)
				if err != nil {
					return
				}

				buf := make([]byte, size)
				_, err = io.ReadFull(pipe2, buf[0:])
				if err != nil {
					return
				}

				err = wsutil.WriteServerBinary(conn, buf)
				if err != nil {
					return
				}
			}
		}()

		// wait for grs to be done
		wg.Wait()

	}))

	return err
}
