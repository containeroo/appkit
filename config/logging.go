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

// RegisterLogFormat adds the shared log-format flag without server-specific logging flags.
func RegisterLogFormat(tf *tinyflags.FlagSet, format *logging.LogFormat) func() {
	return registerLogFormat(tf, format, "")
}

// RegisterLogging adds logging flags, returning a finalizer for the enum value.
func RegisterLogging(tf *tinyflags.FlagSet, cfg *Logging) func() {
	finalizeFormat := registerLogFormat(tf, &cfg.Format, SectionLogging)
	RegisterDebug(tf, cfg)
	RegisterAccessLog(tf, cfg)
	return finalizeFormat
}

// RegisterDebug adds the diagnostic logging flag.
func RegisterDebug(tf *tinyflags.FlagSet, cfg *Logging) {
	tf.BoolVar(&cfg.Debug, "debug", false, "Enable verbose diagnostic logging").
		Section(SectionLogging).
		Short("d").
		Value()
}

// RegisterAccessLog adds the HTTP request logging flag.
func RegisterAccessLog(tf *tinyflags.FlagSet, cfg *Logging) {
	tf.BoolVar(&cfg.AccessLog, "access-log", false, "Enable HTTP request logging").
		Section(SectionLogging).
		Value()
}

// registerLogFormat adds the shared enum and optionally places it in a flag section.
func registerLogFormat(tf *tinyflags.FlagSet, format *logging.LogFormat, section string) func() {
	flag := tinyflags.Enum(tf, "log-format", logging.LogFormatJSON, "Log output format",
		logging.LogFormatText, logging.LogFormatJSON).
		Short("l").
		Placeholder("FORMAT")
	if section != "" {
		flag.Section(section)
	}
	value := flag.Value()
	return func() { *format = *value }
}
