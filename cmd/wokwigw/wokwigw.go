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

	"github.com/wokwi/wokwigw/pkg/loopback"

	"github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/containers/gvisor-tap-vsock/pkg/virtualnetwork"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

const (
	version   = "0.0.1"
	port      = 9011
	gatewayIP = "10.13.37.1"

	/* Forwarding */
	forwardPort   = 9080
	forwardTarget = "10.13.37.2:80"
)

func main() {
	fmt.Printf("Wokwi Gateway version %s\n\n", version)
	fmt.Printf("Listening on TCP port %d\n", port)

	ctx, _ := context.WithCancel(context.Background())

	config := types.Configuration{
		Debug:             false,
		CaptureFile:       "",
		MTU:               1500,
		Subnet:            "10.13.37.0/24",
		GatewayIP:         gatewayIP,
		GatewayMacAddress: "42:13:37:55:aa:01",
		DHCPStaticLeases: map[string]string{
			"10.13.37.2": "24:0a:c4:00:01:10",
		},
		DNS: []types.Zone{
			{
				Name: "wokwi.internal.",
				Records: []types.Record{
					{
						Name: "gateway",
						IP:   net.ParseIP(gatewayIP),
					},
					{
						Name: "host",
						IP:   net.ParseIP("10.13.37.254"),
					},
				},
			},
		},
		Forwards: map[string]string{
			fmt.Sprintf(":%d", forwardPort): forwardTarget,
		},
		NAT: map[string]string{
			"10.13.37.254": "127.0.0.1",
		},
		GatewayVirtualIPs: []string{"10.13.37.254"},
		Protocol:          types.QemuProtocol,
	}

	vn, err := virtualnetwork.New(&config)
	if err != nil {
		fmt.Printf("Error creating network: %v\n", err)
		return
	}

	http.ListenAndServe(":9011", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Client connected: %s\n", r.RemoteAddr)

		origin := r.Header.Get("Origin")
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

		cleanup := func() {
			conn.Close()
			pipe1.Close()
			pipe2.Close()
		}

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
	}))
}
