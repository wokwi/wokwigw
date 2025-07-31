// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2025 Uri Shaked <uri@wokwi.com>

package main

import (
	"context"
	"net"
)

type Backend interface {
	Setup(ctx context.Context) error
	HandleConnection(ctx context.Context, conn net.Conn, remoteAddr string) error
	Cleanup() error
}
