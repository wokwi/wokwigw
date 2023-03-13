package main

import (
	"fmt"
	"net"

	"github.com/containers/gvisor-tap-vsock/pkg/types"
)

const (
	defaultListenPort     = 9011
	defaultListenAddr     = "127.0.0.1"
	defaultHostAddr       = "10.13.37.254"
	defaultGatewayAddr    = "10.13.37.1"
	defaultGatewayMACAddr = "42:13:37:55:aa:01"

	/* Forwarding */
	defaultForwardPort = 9080
	defaultForwardAddr = "10.13.37.2:80"
	defaultSubnet      = "10.13.37.0/24"
)

type flagCfg struct {
	forwardList []string
	listenPort  int
	captureFile string
}

func defaultConfig() types.Configuration {
	return types.Configuration{
		Debug:             false,
		CaptureFile:       "",
		MTU:               1500,
		Subnet:            defaultSubnet,
		GatewayIP:         defaultGatewayAddr,
		GatewayMacAddress: defaultGatewayMACAddr,
		DHCPStaticLeases: map[string]string{
			"10.13.37.2": "24:0a:c4:00:01:10",
		},
		DNS: []types.Zone{
			{
				Name: "wokwi.internal.",
				Records: []types.Record{
					{
						Name: "gateway",
						IP:   net.ParseIP(defaultGatewayAddr),
					},
					{
						Name: "host",
						IP:   net.ParseIP(defaultHostAddr),
					},
				},
			},
		},
		Forwards: map[string]string{
			fmt.Sprintf(":%d", defaultForwardPort): defaultForwardAddr,
		},
		NAT: map[string]string{
			defaultHostAddr: defaultListenAddr,
		},
		GatewayVirtualIPs: []string{defaultHostAddr},
		Protocol:          types.QemuProtocol,
	}
}
