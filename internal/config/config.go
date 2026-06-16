package config

import "flag"

const (
	DefaultFrontendAddr = "127.0.0.1:8080"
	DefaultBackendAddr  = "127.0.0.1:8081"
	DefaultBackendURL   = "http://127.0.0.1:8081"
	DefaultDBPath       = "rayboard.sqlite"
)

type Config struct {
	FrontendAddr string
	BackendAddr  string
	BackendURL   string
	DBPath       string
}

func Default() Config {
	return Config{
		FrontendAddr: DefaultFrontendAddr,
		BackendAddr:  DefaultBackendAddr,
		BackendURL:   DefaultBackendURL,
		DBPath:       DefaultDBPath,
	}
}

func (c *Config) BindRuntimeFlags(flags *flag.FlagSet) {
	flags.StringVar(&c.FrontendAddr, "frontend-addr", c.FrontendAddr, "frontend server address")
	flags.StringVar(&c.BackendAddr, "backend-addr", c.BackendAddr, "backend API server address")
	flags.StringVar(&c.BackendURL, "backend-url", c.BackendURL, "backend API base URL used by the frontend")
	flags.StringVar(&c.DBPath, "db", c.DBPath, "SQLite database path")
}
