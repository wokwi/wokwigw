// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2025 Uri Shaked <uri@wokwi.com>

package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
	"github.com/songgao/packets/ethernet"
	"github.com/songgao/water"
)

type WaterBackend struct {
	ifce       *water.Interface
	config     *types.Configuration
	pcapWriter *pcapgo.Writer
	pcapFile   *os.File
}

func NewWaterBackend(config *types.Configuration) *WaterBackend {
	return &WaterBackend{
		config: config,
	}
}

func (w *WaterBackend) Setup(ctx context.Context) error {
	ifce, err := water.New(water.Config{
		DeviceType: water.TAP,
	})
	if err != nil {
		return fmt.Errorf("error creating TAP interface: %w", err)
	}
	fmt.Println("Created a TAP interface with name:", ifce.Name())
	w.ifce = ifce

	// Setup PCAP file writing if capture file is specified
	if w.config.CaptureFile != "" {
		if err := w.setupPCAP(); err != nil {
			return fmt.Errorf("error setting up PCAP file: %w", err)
		}
		fmt.Printf("PCAP capture enabled: %s\n", w.config.CaptureFile)
	}

	return nil
}

func (w *WaterBackend) setupPCAP() error {
	file, err := os.Create(w.config.CaptureFile)
	if err != nil {
		return fmt.Errorf("error creating PCAP file: %w", err)
	}

	writer := pcapgo.NewWriter(file)
	err = writer.WriteFileHeader(65536, layers.LinkTypeEthernet)
	if err != nil {
		file.Close()
		return fmt.Errorf("error writing PCAP header: %w", err)
	}

	w.pcapFile = file
	w.pcapWriter = writer
	return nil
}

func (w *WaterBackend) writePCAP(data []byte, inbound bool) {
	if w.pcapWriter == nil {
		return
	}

	ci := gopacket.CaptureInfo{
		Timestamp:     time.Now(),
		CaptureLength: len(data),
		Length:        len(data),
	}

	err := w.pcapWriter.WritePacket(ci, data)
	if err != nil {
		fmt.Printf("Error writing to PCAP file: %v\n", err)
	}
}

func (w *WaterBackend) HandleConnection(ctx context.Context, conn net.Conn, remoteAddr string) error {
	return handleWebSocketWithTAP(ctx, conn, w.ifce, remoteAddr, w)
}

func (w *WaterBackend) Cleanup() error {
	if w.pcapFile != nil {
		w.pcapFile.Close()
	}
	if w.ifce != nil {
		return w.ifce.Close()
	}
	return nil
}

func handleWebSocketWithTAP(ctx context.Context, conn net.Conn, ifce *water.Interface, remoteAddr string, backend *WaterBackend) error {
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

			backend.writePCAP(frame[:n], true)

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
				backend.writePCAP(msg, false)

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
