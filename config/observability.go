package config

import (
	"errors"
	"strings"

	"github.com/containeroo/tinyflags"
)

// Observability holds optional application monitoring switches.
type Observability struct {
	// DisableMetrics disables metrics and HTTP instrumentation.
	DisableMetrics bool
	// PerformanceDiagnostics enables extra request and browser measurements.
	PerformanceDiagnostics bool
}

// RegisterObservability adds optional metrics and diagnostics flags.
func RegisterObservability(tf *tinyflags.FlagSet, cfg *Observability) {
	tf.BoolVar(&cfg.DisableMetrics, "disable-metrics", false, "Disable metrics and request instrumentation").
		Section(SectionLogging).
		Value()

	tf.BoolVar(&cfg.PerformanceDiagnostics, "performance-diagnostics", false,
		"Enable request and browser performance diagnostics").
		Section(SectionLogging).
		Value()
}

// OpenTelemetry holds optional OTLP trace-export settings.
type OpenTelemetry struct {
	// ServiceName is the logical service name displayed in traces.
	ServiceName string
	// Endpoint is the OTLP gRPC collector address.
	Endpoint string
	// Insecure permits plaintext gRPC collector transport.
	Insecure bool
}

// RegisterOpenTelemetry adds optional OTLP exporter flags.
func RegisterOpenTelemetry(tf *tinyflags.FlagSet, cfg *OpenTelemetry, defaultServiceName string) {
	tf.StringVar(&cfg.ServiceName, "otel-service-name", defaultServiceName, "OpenTelemetry service name").
		Section(SectionTelemetry).
		Placeholder("NAME").
		Validate(func(value string) error {
			if strings.TrimSpace(value) == "" {
				return errors.New("OpenTelemetry service name is required")
			}
			return nil
		}).
		Value()

	tf.StringVar(&cfg.Endpoint, "otel-exporter-otlp-endpoint", "", "OTLP gRPC exporter endpoint").
		Section(SectionTelemetry).
		Placeholder("HOST:PORT").
		Value()

	tf.BoolVar(&cfg.Insecure, "otel-exporter-otlp-insecure", false,
		"Use insecure OTLP gRPC exporter transport").
		Section(SectionTelemetry).
		Requires("otel-exporter-otlp-endpoint").
		Value()
}
