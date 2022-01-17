// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2022 Uri Shaked <uri@wokwi.com>

package main

type alohaMessage struct {
	Type           string `json:"type"`
	Protocol       string `json:"protocol"`
	Version        int32  `json:"version"`
	GatewayVersion string `json:"gatewayVersion"`
}

func makeAlohaMessage(version string) alohaMessage {
	return alohaMessage{
		Type:           "aloha",
		Protocol:       "wokwigw",
		Version:        1,
		GatewayVersion: version,
	}
}
