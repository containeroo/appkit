package config

import (
	"github.com/containeroo/appkit/logging"
	"github.com/containeroo/tinyflags"
)

// Logging holds the shared application logging flag values.
type Logging struct {
	// Format chooses JSON or structured text output.
	Format logging.LogFormat
	// Debug enables debug-level events.
	Debug bool
	// AccessLog enables per-request logging.
	AccessLog bool
}

// RegisterLogging adds logging flags, returning a finalizer for the enum value.
func RegisterLogging(tf *tinyflags.FlagSet, cfg *Logging) func() {
	format := tinyflags.Enum(tf, "log-format", logging.LogFormatJSON, "Log output format",
		logging.LogFormatText, logging.LogFormatJSON).
		Section(SectionLogging).
		Short("l").Placeholder("FORMAT").
		Value()

	tf.BoolVar(&cfg.Debug, "debug", false, "Enable verbose diagnostic logging").
		Section(SectionLogging).
		Short("d").
		Value()

	tf.BoolVar(&cfg.AccessLog, "access-log", false, "Enable HTTP request logging").
		Section(SectionLogging).
		Value()
	return func() { cfg.Format = *format }
}
