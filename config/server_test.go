package config

import (
	"net"
	"testing"
)

func TestRegisterServerAddresses(t *testing.T) {
	cases := []struct {
		name       string
		opts       ServerOptions
		args       []string
		listen     string
		management string
	}{
		{
			name:       "defaults",
			listen:     "127.0.0.1:8080",
			management: "127.0.0.1:8081",
		},
		{
			name:       "overridden listeners",
			args:       []string{"--listen-address", "127.0.0.1:9010", "--management-listen-address", "127.0.0.1:9011"},
			listen:     "127.0.0.1:9010",
			management: "127.0.0.1:9011",
		},
		{
			name: "custom defaults",
			opts: ServerOptions{
				ListenAddress:           &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9090},
				ManagementListenAddress: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9091},
			},
			listen:     "127.0.0.1:9090",
			management: "127.0.0.1:9091",
		},
		{
			name:   "without management",
			opts:   ServerOptions{WithoutManagement: true},
			listen: "127.0.0.1:8080",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			flags := NewFlagSet("example", "EXAMPLE", "dev")
			var cfg Server
			finalize := RegisterServer(flags, &cfg, test.opts)
			if err := flags.Parse(test.args); err != nil {
				t.Fatal(err)
			}
			finalize()
			if cfg.ListenAddress != test.listen {
				t.Errorf("ListenAddress = %q, want %q", cfg.ListenAddress, test.listen)
			}
			if cfg.ManagementListenAddress != test.management {
				t.Errorf("ManagementListenAddress = %q, want %q", cfg.ManagementListenAddress, test.management)
			}
		})
	}
}
