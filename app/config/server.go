package config

// ServerConfig - configs for the server
type ServerConfig struct {
	Version    string `yaml:"version"`
	RuntimeDir string `yaml:"runtime_dir"`
	LogDir     string `yaml:"log_dir"`
	PublicDir  string `yaml:"public_dir"`
}
