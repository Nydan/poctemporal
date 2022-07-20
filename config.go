package poctemporal

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	TLS      TLS      `yaml:"tls"`
	Temporal Temporal `yaml:"temporal"`
}

type TLS struct {
	CertPoolPath string `yaml:"cert-pool-path"`
	CertPath     string `yaml:"cert-path"`
	KeyPath      string `yaml:"key-path"`
}

type Temporal struct {
	HostPort   string `yaml:"hostport"`
	Namespace  string `yaml:"namespace"`
	ServerName string `yaml:"server-name"`
}

func Load(path string) (Config, error) {
	var cfg Config

	info, err := os.Stat(path)
	if err != nil {
		return cfg, err
	}

	if info.IsDir() {
		return cfg, fmt.Errorf("%s is a directory", path)
	}

	// We need to create a new clean path variable for gosec.
	cleanPath := filepath.Clean(path)
	out, err := os.ReadFile(cleanPath)
	if err != nil {
		return cfg, err
	}

	if err := yaml.Unmarshal(out, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
