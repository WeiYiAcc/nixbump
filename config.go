package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Source       string   `json:"source" yaml:"source"`
	Owner        string   `json:"owner,omitempty" yaml:"owner,omitempty"`
	Repo         string   `json:"repo,omitempty" yaml:"repo,omitempty"`
	Host         string   `json:"host,omitempty" yaml:"host,omitempty"`
	Package      string   `json:"package,omitempty" yaml:"package,omitempty"`
	StripPrefix  string   `json:"strip_prefix,omitempty" yaml:"strip_prefix,omitempty"`
	VersionRegex string   `json:"version_regex,omitempty" yaml:"version_regex,omitempty"`
	ExtractHash  string   `json:"extract_hash,omitempty" yaml:"extract_hash,omitempty"`
	Args         []string `json:"args,omitempty" yaml:"args,omitempty"`
}

var validSources = map[string]bool{
	"github":     true,
	"gitlab":     true,
	"npm":        true,
	"nix-update": true,
}

// sopsMetadataRe matches a top-level "sops:" (yaml) or '"sops":' (json) key,
// which marks a file encrypted by sops (field-level metadata block).
var sopsMetadataRe = regexp.MustCompile(`(?m)^\s*"?sops"?\s*:`)

func parseConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	// sops-encrypted config (contains token fields): decrypt via sops CLI.
	// Piping sops output is safe (no disk, no shell history) per repo convention.
	if sopsMetadataRe.Match(data) {
		decrypted, err := sopsDecrypt(data)
		if err != nil {
			return nil, fmt.Errorf("sops decrypt %s: %w", path, err)
		}
		data = decrypted
	}

	var cfg Config
	switch strings.ToLower(filepath.Ext(path)) {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parsing config: %w", err)
		}
	default:
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parsing config: %w", err)
		}
	}

	if !validSources[cfg.Source] {
		return nil, fmt.Errorf("unknown source: %q", cfg.Source)
	}

	if cfg.Source == "gitlab" && cfg.Host == "" {
		cfg.Host = "gitlab.com"
	}

	return &cfg, nil
}

func sopsDecrypt(data []byte) ([]byte, error) {
	cmd := exec.Command("sops", "-d", "/dev/stdin")
	cmd.Stdin = bytes.NewReader(data)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%v: %s", err, bytes.TrimSpace(stderr.Bytes()))
	}
	return out, nil
}
