package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFixture(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParseConfig(t *testing.T) {
	path := writeFixture(t, "nixbump.json", `{
		"source": "github",
		"owner": "NixOS",
		"repo": "nixpkgs",
		"strip_prefix": "v",
		"version_regex": "^v(\\d+\\.\\d+)$",
		"extract_hash": "src"
	}`)

	cfg, err := parseConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Source != "github" {
		t.Errorf("source = %q, want %q", cfg.Source, "github")
	}
	if cfg.Owner != "NixOS" {
		t.Errorf("owner = %q, want %q", cfg.Owner, "NixOS")
	}
	if cfg.Repo != "nixpkgs" {
		t.Errorf("repo = %q, want %q", cfg.Repo, "nixpkgs")
	}
	if cfg.StripPrefix != "v" {
		t.Errorf("strip_prefix = %q, want %q", cfg.StripPrefix, "v")
	}
	if cfg.VersionRegex != `^v(\d+\.\d+)$` {
		t.Errorf("version_regex = %q, want %q", cfg.VersionRegex, `^v(\d+\.\d+)$`)
	}
	if cfg.ExtractHash != "src" {
		t.Errorf("extract_hash = %q, want %q", cfg.ExtractHash, "src")
	}
}

func TestParseConfigNixUpdate(t *testing.T) {
	path := writeFixture(t, "nixbump.json", `{
		"source": "nix-update",
		"args": ["--flake", "--version=branch"]
	}`)

	cfg, err := parseConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Source != "nix-update" {
		t.Errorf("source = %q, want %q", cfg.Source, "nix-update")
	}
	if len(cfg.Args) != 2 || cfg.Args[0] != "--flake" || cfg.Args[1] != "--version=branch" {
		t.Errorf("args = %v, want [--flake --version=branch]", cfg.Args)
	}
}

func TestParseConfigInvalidSource(t *testing.T) {
	path := writeFixture(t, "nixbump.json", `{
		"source": "bitbucket"
	}`)

	_, err := parseConfig(path)
	if err == nil {
		t.Fatal("expected error for invalid source, got nil")
	}
}

func TestParseConfigGitLabDefaultHost(t *testing.T) {
	path := writeFixture(t, "nixbump.json", `{
		"source": "gitlab",
		"owner": "inkscape",
		"repo": "inkscape"
	}`)

	cfg, err := parseConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Host != "gitlab.com" {
		t.Errorf("host = %q, want %q", cfg.Host, "gitlab.com")
	}
}
