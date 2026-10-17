package config

import "github.com/containeroo/tinyflags"

// Postgres holds application PostgreSQL deployment settings.
type Postgres struct {
	// URL is the database DSN; secrets must not be printed in plaintext.
	URL string
	// MaxConns is the maximum connection count; zero uses driver defaults.
	MaxConns int32
	// MinIdleConns is the idle connection minimum; zero uses driver defaults.
	MinIdleConns int32
}

// PostgresOptions permits services without mandatory databases to opt out of --database-url.
type PostgresOptions struct {
	// OptionalURL allows the database URL to be omitted.
	OptionalURL bool
}

// RegisterPostgres registers the database URL and common pool settings.
func RegisterPostgres(tf *tinyflags.FlagSet, cfg *Postgres, opts PostgresOptions) {
	RegisterDatabaseURL(tf, cfg, opts)
	RegisterDatabaseMaxConns(tf, cfg)
	RegisterDatabaseMinIdleConns(tf, cfg)
}

// RegisterDatabaseURL adds the PostgreSQL connection URL flag.
func RegisterDatabaseURL(tf *tinyflags.FlagSet, cfg *Postgres, opts PostgresOptions) {
	url := tf.StringVar(&cfg.URL, "database-url", "", "PostgreSQL connection URL").
		Section(SectionDatabase).
		Placeholder("URL").
		OverriddenValueMaskFn(tinyflags.MaskPostgresURL)
	if !opts.OptionalURL {
		url.Required()
	}
	url.Value()
}

// RegisterDatabaseMaxConns adds the maximum connection count flag.
func RegisterDatabaseMaxConns(tf *tinyflags.FlagSet, cfg *Postgres) {
	tf.Int32Var(&cfg.MaxConns, "database-max-conns", 0, "Maximum database connections; 0 uses driver defaults").
		Section(SectionDatabase).
		Validate(tinyflags.NonNegative[int32]()).
		Value()
}

// RegisterDatabaseMinIdleConns adds the minimum idle connection count flag.
func RegisterDatabaseMinIdleConns(tf *tinyflags.FlagSet, cfg *Postgres) {
	tf.Int32Var(&cfg.MinIdleConns, "database-min-idle-conns", 0, "Minimum idle database connections; 0 uses driver defaults").
		Section(SectionDatabase).
		Validate(tinyflags.NonNegative[int32]()).
		Value()
}
