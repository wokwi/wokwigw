// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2022 Uri Shaked <uri@wokwi.com>

package main

import (
	"testing"
)

func TestCheckOrigin(t *testing.T) {
	tests := []struct {
		in       string
		expected bool
	}{
		{"https://wokwi.com/", true},
		{"https://Wokwi.com/", true},
		{"https://something.preview.Wokwi.com/", true},
		{"http://localhost:3000/", true},
		{"http://127.0.0.1/", true},
		{"https://invalid.wokwi.com/", false},
		{"https://127.0.0.1/", false},
		{"http://notwokwi.com/", false},
		{"https://notwokwi.com/", false},
		{"invalid url", false},
	}

	for _, testCase := range tests {
		actual := checkOrigin(testCase.in)
		if actual != testCase.expected {
			t.Errorf("should return %v instead of %v given %s", testCase.expected, actual, testCase.in)
		}
	}
}
