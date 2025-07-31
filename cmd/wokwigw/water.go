// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2025 Uri Shaked <uri@wokwi.com>

package main

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/songgao/packets/ethernet"
	"github.com/songgao/water"
)

type WaterBackend struct {
	ifce *water.Interface
}

func NewWaterBackend() *WaterBackend {
	return &WaterBackend{}
}

func (w *WaterBackend) Setup(ctx context.Context) error {
	ifce, err := water.New(water.Config{
		DeviceType: water.TAP,
	})
	if err != nil {
		return fmt.Errorf("error creating TAP interface: %w", err)
	}
	w.ifce = ifce
	return nil
}

func (w *WaterBackend) HandleConnection(ctx context.Context, conn net.Conn, remoteAddr string) error {
	return handleWebSocketWithTAP(ctx, conn, w.ifce, remoteAddr)
}

func (w *WaterBackend) Cleanup() error {
	if w.ifce != nil {
		return w.ifce.Close()
	}
	return nil
}

func handleWebSocketWithTAP(ctx context.Context, conn net.Conn, ifce *water.Interface, remoteAddr string) error {
	wg := sync.WaitGroup{}
	_, cancel := context.WithCancel(ctx)

	wg.Add(2)
	cleanup := func() {
		cancel()
		_ = conn.Close()
		wg.Done()
	}

	// Read from TAP interface and send to websocket
	go func() {
		defer cleanup()

		var frame ethernet.Frame
		frame.Resize(1500)

		for {
			n, err := ifce.Read([]byte(frame))
			if err != nil {
				return
			}

			err = wsutil.WriteServerBinary(conn, frame[:n])
			if err != nil {
				return
			}
		}
	}()

	// Read from websocket and write to TAP interface
	go func() {
		defer cleanup()

		for {
			msg, op, err := wsutil.ReadClientData(conn)
			if err != nil {
				return
			}
			switch op {
			case ws.OpBinary:
				_, err = ifce.Write(msg)
				if err != nil {
					return
				}
			case ws.OpText:
				fmt.Printf("[%s] Incoming message: %v\n", remoteAddr, msg)
			}
		}
	}()

	wg.Wait()
	return nil
}
