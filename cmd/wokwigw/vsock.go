// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2022-2025 Uri Shaked <uri@wokwi.com>

package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/containers/gvisor-tap-vsock/pkg/virtualnetwork"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/wokwi/wokwigw/pkg/loopback"
)

type VsockBackend struct {
	config *types.Configuration
	vn     *virtualnetwork.VirtualNetwork
}

func NewVsockBackend(config *types.Configuration) *VsockBackend {
	return &VsockBackend{
		config: config,
	}
}

func (v *VsockBackend) Setup(ctx context.Context) error {
	vn, err := virtualnetwork.New(v.config)
	if err != nil {
		return fmt.Errorf("error creating network %w", err)
	}
	v.vn = vn
	return nil
}

func (v *VsockBackend) HandleConnection(ctx context.Context, conn net.Conn, remoteAddr string) error {
	pipe1, pipe2, err := loopback.ConnLoopback()
	if err != nil {
		return fmt.Errorf("pipe creation failed: %w", err)
	}
	defer pipe1.Close()
	defer pipe2.Close()

	go v.vn.AcceptQemu(ctx, pipe1)

	return handleWebSocketCommunication(ctx, conn, pipe2, remoteAddr)
}

func (v *VsockBackend) Cleanup() error {
	return nil
}

func handleWebSocketCommunication(ctx context.Context, conn net.Conn, pipe net.Conn, remoteAddr string) error {
	wg := sync.WaitGroup{}
	_, cancel := context.WithCancel(ctx)

	wg.Add(2)
	cleanup := func() {
		cancel()
		// todo: need to handle errors here
		_ = conn.Close()
		_ = pipe.Close()
		wg.Done()
	}

	go func() {
		defer cleanup()

		for {
			msg, op, err := wsutil.ReadClientData(conn)
			if err != nil {
				return
			}
			switch op {
			case ws.OpBinary:
				err := binary.Write(pipe, binary.BigEndian, uint32(len(msg)))
				if err != nil {
					return
				}

				_, err = pipe.Write(msg)
				if err != nil {
					return
				}

			case ws.OpText:
				fmt.Printf("[%s] Incoming message: %v\n", remoteAddr, msg)
			}
		}
	}()

	go func() {
		defer cleanup()

		for {
			var size uint32
			err := binary.Read(pipe, binary.BigEndian, &size)
			if err != nil {
				return
			}

			buf := make([]byte, size)
			_, err = io.ReadFull(pipe, buf[0:])
			if err != nil {
				return
			}

			err = wsutil.WriteServerBinary(conn, buf)
			if err != nil {
				return
			}
		}
	}()

	wg.Wait()
	return nil
}
