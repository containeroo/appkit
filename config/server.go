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
	RegisterRoutePrefix(tf, cfg)
	finishListen := RegisterListenAddress(tf, cfg, opts.ListenAddress)
	var finishManagement func()
	if !opts.WithoutManagement {
		finishManagement = RegisterManagementListenAddress(tf, cfg, opts.ManagementListenAddress)
	}
	RegisterPublicURL(tf, cfg, opts.PublicURL)
	return func() {
		finishListen()
		if finishManagement != nil {
			finishManagement()
		}
	}
}

// RegisterRoutePrefix adds the normalized route prefix flag.
func RegisterRoutePrefix(tf *tinyflags.FlagSet, cfg *Server) {
	tf.StringVar(&cfg.RoutePrefix, "route-prefix", "", "URL path prefix under which the application is served").
		Section(SectionServer).
		Validate(prefix.ValidateRoutePrefix).
		Finalize(prefix.NormalizeRoutePrefix).
		Value()
}

// RegisterListenAddress adds the application listener flag.
// Call the returned finalizer after parsing.
func RegisterListenAddress(tf *tinyflags.FlagSet, cfg *Server, listenDefault *net.TCPAddr) func() {
	if listenDefault == nil {
		listenDefault = &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}
	}
	listen := tf.TCPAddr("listen-address", listenDefault, "Address on which the application server listens").
		Section(SectionServer).
		Short("a").
		Placeholder("ADDR").
		Value()

	return func() { cfg.ListenAddress = (*listen).String() }
}

// RegisterManagementListenAddress adds the management listener flag.
// Call the returned finalizer after parsing.
func RegisterManagementListenAddress(tf *tinyflags.FlagSet, cfg *Server, managementDefault *net.TCPAddr) func() {
	if managementDefault == nil {
		managementDefault = &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8081}
	}
	management := tf.TCPAddr("management-listen-address", managementDefault, "Address on which operational endpoints listen").
		Section(SectionServer).
		Short("m").
		Placeholder("ADDR").
		Value()

	return func() { cfg.ManagementListenAddress = (*management).String() }
}

// RegisterPublicURL adds the externally visible application URL flag.
func RegisterPublicURL(tf *tinyflags.FlagSet, cfg *Server, publicURL string) {
	if publicURL == "" {
		publicURL = "http://localhost:8080"
	}
	tf.StringVar(&cfg.PublicURL, "public-url", publicURL, "Externally visible application URL").
		Section(SectionServer).
		Placeholder("URL").
		Value()

}
