package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Source       string   `json:"source"`
	Owner        string   `json:"owner,omitempty"`
	Repo         string   `json:"repo,omitempty"`
	Host         string   `json:"host,omitempty"`
	Package      string   `json:"package,omitempty"`
	StripPrefix  string   `json:"strip_prefix,omitempty"`
	VersionRegex string   `json:"version_regex,omitempty"`
	ExtractHash  string   `json:"extract_hash,omitempty"`
	Args         []string `json:"args,omitempty"`
}

var validSources = map[string]bool{
	"github":     true,
	"gitlab":     true,
	"npm":        true,
	"nix-update": true,
}

func parseConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if !validSources[cfg.Source] {
		return nil, fmt.Errorf("unknown source: %q", cfg.Source)
	}

	if cfg.Source == "gitlab" && cfg.Host == "" {
		cfg.Host = "gitlab.com"
	}

	return &cfg, nil
}
