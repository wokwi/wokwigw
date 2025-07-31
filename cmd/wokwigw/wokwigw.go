// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2022 Uri Shaked <uri@wokwi.com>

package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/sirupsen/logrus"

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

	f.StringSliceVar(&flags.forwardList, "forward", flags.forwardList, "forward port to the simulator. Format: [udp:]localPort:remoteAddress:remotePort tuples")
	f.IntVar(&flags.listenPort, "listenPort", flags.listenPort, "listening port (on localhost)")
	f.StringVar(&flags.captureFile, "captureFile", flags.captureFile, "packet capture (PCAP) file name (for debugging)")
	f.BoolVar(&flags.bridge, "bridge", flags.bridge, "use bridge mode (experimental, see docs)")

	return rootCmd
}

func validateAndMapFlags(flags *flagCfg, cfg *types.Configuration) (err error) {

	// Check if bridge mode is incompatible with forwards
	if flags.bridge && len(flags.forwardList) > 0 {
		return fmt.Errorf("bridge mode does not support port forwarding. remove the --forward flag")
	}

	// map command line arguments to configuration structure (forwardList)
	// note: since we're using syntax similar to ssh -L option, we do the splitting ourselves here
	for _, fwd := range flags.forwardList {
		parts := strings.Split(fwd, ":")
		prefix := ""
		if len(parts) == 4 && parts[0] == "udp" {
			prefix = "udp:"
			parts = parts[1:]
		}
		if len(parts) != 3 {
			return fmt.Errorf(string("arg ``%s`` is not formatted using the syntax '[udp:]localPort:addr:remotePort'"), fwd)
		}

		if v, e := strconv.Atoi(parts[0]); e != nil || v < 0 || v > 65535 {
			return fmt.Errorf("invalid local port specified in forward argument (%s): %w", fwd, e)
		}

		if v, e := strconv.Atoi(parts[2]); e != nil || v < 0 || v > 65535 {
			return fmt.Errorf("invalid remote port specified in forward argument (%s): %w", fwd, e)
		}

		cfg.Forwards[prefix+":"+parts[0]] = net.JoinHostPort(parts[1], parts[2])
	}

	if flags.listenPort < 0 || flags.listenPort > 65535 {
		return fmt.Errorf("invalid listen port specified (%d)", flags.listenPort)
	}

	cfg.CaptureFile = flags.captureFile

	return nil
}

func banner() {

	var gitStr string

	if gitHash != "" {
		gitStr = fmt.Sprintf("Git revision: %s\nBuilt: %s\n\n", gitHash, buildTime)
	}

	mode := "vsock"
	if flags.bridge {
		mode = "bridge"
	}
	fmt.Printf(`
       __              ,
|  |  /  \  |_/  |  |  |
|/\|  \__/  | \  |/\|  |

    Wokwi IoT Gateway

Version: %s
%s
Listening on TCP Port %d (mode: %s)
`, version, gitStr, flags.listenPort, mode)
}

func printForwards(config *types.Configuration) {
	// Print forwards
	if len(config.Forwards) > 0 {
		fmt.Println("\nPort forwards (local -> simulator):")
		for local, remote := range config.Forwards {
			fmt.Printf("  %s -> %s\n", local, remote)
		}
		fmt.Println()
	}
}

func run(cmd *cobra.Command, _ []string) error {
	logrus.SetLevel(logrus.WarnLevel)

	banner()

	// Create the appropriate backend based on the bridge flag
	var backend Backend
	if flags.bridge {
		backend = NewWaterBackend()
	} else {
		printForwards(&config)
		backend = NewVsockBackend(&config)
	}

	// Setup the backend
	ctx := cmd.Context()
	if err := backend.Setup(ctx); err != nil {
		return fmt.Errorf("error setting up backend: %w", err)
	}
	defer backend.Cleanup()

	err := http.ListenAndServe(net.JoinHostPort(defaultListenAddr, strconv.Itoa(flags.listenPort)), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		// Handle the connection using the appropriate backend
		if err := backend.HandleConnection(ctx, conn, r.RemoteAddr); err != nil {
			fmt.Printf("[%s] Connection handling error: %s\n", r.RemoteAddr, err)
		}
	}))

	return err
}
