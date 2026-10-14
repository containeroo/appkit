package config

import (
	"net"

	"github.com/containeroo/httpkit/prefix"
	"github.com/containeroo/tinyflags"
)

// Server is shared deployment-level HTTP listener configuration.
type Server struct {
	// RoutePrefix is the normalized application deployment prefix.
	RoutePrefix string
	// ListenAddress is the application TCP listen address.
	ListenAddress string
	// ManagementListenAddress is the management TCP listen address, when enabled.
	ManagementListenAddress string
	// PublicURL is the externally reachable application URL.
	PublicURL string
}

// ServerOptions controls application-specific listener defaults and optional flags.
type ServerOptions struct {
	// ListenAddress defaults to 127.0.0.1:8080 when nil.
	ListenAddress *net.TCPAddr
	// ManagementListenAddress defaults to 127.0.0.1:8081 when nil.
	ManagementListenAddress *net.TCPAddr
	// PublicURL defaults to http://localhost:8080 when empty.
	PublicURL string
	// WithoutManagement omits the management-listen-address flag.
	WithoutManagement bool
}

// RegisterServer adds server flags and returns a finalizer to call after Parse.
// The application remains responsible for prefix-aware handler mounting.
func RegisterServer(tf *tinyflags.FlagSet, cfg *Server, opts ServerOptions) func() {
	tf.StringVar(&cfg.RoutePrefix, "route-prefix", "", "URL path prefix under which the application is served").
		Section(SectionServer).
		Validate(prefix.ValidateRoutePrefix).
		Finalize(prefix.NormalizeRoutePrefix).
		Value()

	listenDefault := opts.ListenAddress
	if listenDefault == nil {
		listenDefault = &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}
	}
	listen := tf.TCPAddr("listen-address", listenDefault, "Address on which the application server listens").
		Section(SectionServer).
		Short("a").
		Placeholder("ADDR").
		Value()

	var management **net.TCPAddr
	if !opts.WithoutManagement {
		managementDefault := opts.ManagementListenAddress
		if managementDefault == nil {
			managementDefault = &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8081}
		}
		management = tf.TCPAddr("management-listen-address", managementDefault, "Address on which operational endpoints listen").
			Section(SectionServer).
			Short("m").
			Placeholder("ADDR").
			Value()
	}

	publicURL := opts.PublicURL
	if publicURL == "" {
		publicURL = "http://localhost:8080"
	}
	tf.StringVar(&cfg.PublicURL, "public-url", publicURL, "Externally visible application URL").
		Section(SectionServer).
		Placeholder("URL").
		Value()

	return func() {
		cfg.ListenAddress = (*listen).String()
		if management != nil {
			cfg.ManagementListenAddress = (*management).String()
		}
	}
}
