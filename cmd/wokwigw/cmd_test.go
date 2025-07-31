package main

import (
	"bytes"
	"testing"

	"github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestFlags(t *testing.T) {

	tcs := map[string]struct {
		fwds       []string
		wantFwdCnt int
		listenPort int
		bridge     bool
		wantErr    bool
		errStrPfx  string
	}{
		"1 valid fwd, no listen port":                 {[]string{"1234:host:4567"}, 1, 9011, false, false, ""},
		"2 valid fwd, no listen port":                 {[]string{"1234:host:4567", "1111:host:2222"}, 2, 9011, false, false, ""},
		"1 invalid fwd (local port), no listen port":  {[]string{"99999:host:4567"}, 0, 9011, false, true, "invalid local port specified in forward argument"},
		"2 invalid fwd (local port), no listen port":  {[]string{"1234:host:1234", "99999:host:4567"}, 0, 9011, false, true, "invalid local port specified in forward argument"},
		"1 invalid fwd (remote port), no listen port": {[]string{"1234:host:99999"}, 0, 9011, false, true, "invalid remote port specified in forward argument"},
		"0 fwd, invalid listen port":                  {[]string{}, 0, 99999, false, true, "invalid listen port specified"},
		"bridge mode with forwards":                   {[]string{"1234:host:4567"}, 0, 9011, true, true, "bridge mode does not support port forwarding"},
		"bridge mode without forwards":                {[]string{}, 0, 9011, true, false, ""},
	}

	for _, tc := range tcs {

		f := flagCfg{
			forwardList: tc.fwds,
			listenPort:  tc.listenPort,
			bridge:      tc.bridge,
		}
		cfg := types.Configuration{
			Forwards: map[string]string{},
		}

		err := validateAndMapFlags(&f, &cfg)
		if !tc.wantErr {
			assert.NoError(t, err)
			assert.Equal(t, tc.wantFwdCnt, len(cfg.Forwards))
			assert.Equal(t, f.listenPort, tc.listenPort)
		} else {
			assert.Contains(t, err.Error(), tc.errStrPfx)
		}

	}

}

func TestRootCmd(t *testing.T) {

	tcs := map[string]struct {
		args       []string
		wantFwdCnt int
		listenPort int
		bridge     bool
		wantErr    bool
		errStrPfx  string
	}{
		"1 valid fwd, no listen port":                 {[]string{"--forward", "1234:host:4567"}, 1, 0, false, false, ""},
		"2 valid fwd, no listen port":                 {[]string{"--forward", "1234:host:4567", "--forward", "1111:host:2222"}, 2, 0, false, false, ""},
		"1 valid fwd with udp":                        {[]string{"--forward", "udp:1234:host:4567"}, 1, 0, false, false, ""},
		"1 invalid fwd (local port), no listen port":  {[]string{"--forward", "99999:host:4567"}, 0, 0, false, true, "invalid local port specified in forward argument"},
		"2 invalid fwd (local port), no listen port":  {[]string{"--forward", "1234:host:1234", "--forward", "99999:host:4567"}, 0, 0, false, true, "invalid local port specified in forward argument"},
		"1 invalid fwd (remote port), no listen port": {[]string{"--forward", "1234:host:99999"}, 0, 9011, false, true, "invalid remote port specified in forward argument"},
		"0 fwd, invalid listen port":                  {[]string{"--listenPort", "99999"}, 0, 99999, false, true, "invalid listen port specified"},
		"bridge mode enabled":                         {[]string{"--bridge"}, 0, 0, true, false, ""},
		"bridge mode with forward":                    {[]string{"--bridge", "--forward", "1234:host:4567"}, 0, 0, true, true, "bridge mode does not support port forwarding"},
		"bridge mode with multiple forwards":          {[]string{"--bridge", "--forward", "1234:host:4567", "--forward", "5678:host:9012"}, 0, 0, true, true, "bridge mode does not support port forwarding"},
		"bridge mode with udp forward":                {[]string{"--bridge", "--forward", "udp:1234:host:4567"}, 0, 0, true, true, "bridge mode does not support port forwarding"},
	}

	for name, tc := range tcs {
		tc := tc
		t.Run(name, func(t *testing.T) {
			f := flagCfg{
				forwardList: []string{},
				listenPort:  0,
			}
			cfg := types.Configuration{
				Forwards: map[string]string{},
			}

			cmd := newRootCmd(&f, &cfg)
			output := &bytes.Buffer{}
			cmd.SetOut(output)
			cmd.SetArgs(tc.args)
			cmd.RunE = func(_ *cobra.Command, _ []string) error { return nil }

			// this will come back immediately, before executing run but after all arg validation
			err := cmd.Execute()

			if !tc.wantErr {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantFwdCnt, len(cfg.Forwards))
				assert.Equal(t, f.listenPort, tc.listenPort)
				assert.Equal(t, tc.bridge, f.bridge)
			} else {
				assert.Contains(t, err.Error(), tc.errStrPfx)
			}
		})

	}
}
