// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2022 Uri Shaked <uri@wokwi.com>

package main

import (
	"net/url"
	"strings"
)

func checkOrigin(origin string) bool {
	originURL, err := url.Parse(origin)
	if err != nil {
		return false
	}

	host := strings.ToLower(originURL.Hostname())
	if originURL.Scheme == "https" {
		return host == "wokwi.com" || strings.HasSuffix(host, ".preview.wokwi.com")
	}

	if originURL.Scheme == "http" {
		return host == "localhost" || host == "127.0.0.1"
	}

	return false
}
