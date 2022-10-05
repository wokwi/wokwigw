package main

import (
	"bytes"
	"github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFlags(t *testing.T) {

	tcs := map[string]struct {
		fwds       []string
		wantFwdCnt int
		listenPort int
		wantErr    bool
		errStrPfx  string
	}{
		"1 valid fwd, no listen port":                 {[]string{"1234:host:4567"}, 1, 9011, false, ""},
		"2 valid fwd, no listen port":                 {[]string{"1234:host:4567", "1111:host:2222"}, 2, 9011, false, ""},
		"1 invalid fwd (local port), no listen port":  {[]string{"99999:host:4567"}, 0, 9011, true, "invalid local port specified in forward argument"},
		"2 invalid fwd (local port), no listen port":  {[]string{"1234:host:1234", "99999:host:4567"}, 0, 9011, true, "invalid local port specified in forward argument"},
		"1 invalid fwd (remote port), no listen port": {[]string{"1234:host:99999"}, 0, 9011, true, "invalid remote port specified in forward argument"},
		"0 fwd, invalid listen port":                  {[]string{}, 0, 99999, true, "invalid listen port specified"},
	}

	for _, tc := range tcs {

		f := flagCfg{
			forwardList: tc.fwds,
			listenPort:  tc.listenPort,
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
		wantErr    bool
		errStrPfx  string
	}{
		"1 valid fwd, no listen port":                 {[]string{"--forward", "1234:host:4567"}, 1, 0, false, ""},
		"2 valid fwd, no listen port":                 {[]string{"--forward", "1234:host:4567", "--forward", "1111:host:2222"}, 2, 0, false, ""},
		"1 invalid fwd (local port), no listen port":  {[]string{"--forward", "99999:host:4567"}, 0, 0, true, "invalid local port specified in forward argument"},
		"2 invalid fwd (local port), no listen port":  {[]string{"--forward", "1234:host:1234", "--forward", "99999:host:4567"}, 0, 0, true, "invalid local port specified in forward argument"},
		"1 invalid fwd (remote port), no listen port": {[]string{"--forward", "1234:host:99999"}, 0, 9011, true, "invalid remote port specified in forward argument"},
		"0 fwd, invalid listen port":                  {[]string{"--listenPort", "99999"}, 0, 99999, true, "invalid listen port specified"},
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
			} else {
				assert.Contains(t, err.Error(), tc.errStrPfx)
			}
		})

	}
}
