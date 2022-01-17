// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2022 Uri Shaked <uri@wokwi.com>

package loopback

import (
	"encoding/binary"
	"testing"
)

func TestConnLoopback(t *testing.T) {
	conn1, conn2, err := ConnLoopback()
	if err != nil {
		t.Error(err)
	}

	var writeVal uint32 = 42
	err = binary.Write(conn1, binary.BigEndian, &writeVal)
	if err != nil {
		t.Error(err)
	}

	var readVal uint32
	err = binary.Read(conn2, binary.BigEndian, &readVal)
	if err != nil {
		t.Error(err)
	}

	if readVal != 42 {
		t.Errorf("Value read (%d) is different from value written (%d)", readVal, writeVal)
	}

	writeVal = 33
	err = binary.Write(conn2, binary.BigEndian, &writeVal)
	if err != nil {
		t.Error(err)
	}

	err = binary.Read(conn1, binary.BigEndian, &readVal)
	if err != nil {
		t.Error(err)
	}

	if readVal != 33 {
		t.Errorf("Value read (%d) is different from value written (%d)", readVal, writeVal)
	}

	conn1.Close()
	conn2.Close()
}
