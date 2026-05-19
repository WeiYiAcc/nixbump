package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverPackages(t *testing.T) {
	root := t.TempDir()
	pkgsDir := filepath.Join(root, "pkgs")

	// Package with nixbump.json
	pkg1 := filepath.Join(pkgsDir, "foo")
	os.MkdirAll(pkg1, 0755)
	os.WriteFile(filepath.Join(pkg1, "nixbump.json"), []byte(`{"source":"github","owner":"o","repo":"r"}`), 0644)
	os.WriteFile(filepath.Join(pkg1, "default.nix"), []byte("{}"), 0644)

	// Package without config (should be skipped)
	pkg2 := filepath.Join(pkgsDir, "bar")
	os.MkdirAll(pkg2, 0755)
	os.WriteFile(filepath.Join(pkg2, "default.nix"), []byte("{}"), 0644)

	// Package with nix-update config
	pkg3 := filepath.Join(pkgsDir, "baz")
	os.MkdirAll(pkg3, 0755)
	os.WriteFile(filepath.Join(pkg3, "nixbump.json"), []byte(`{"source":"nix-update","args":["--version-regex","(.*)"]}`), 0644)

	pkgs, err := discoverPackages(pkgsDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(pkgs) != 2 {
		t.Fatalf("got %d packages, want 2", len(pkgs))
	}
	if pkgs[0].Name != "baz" {
		t.Errorf("first package: got %q, want baz", pkgs[0].Name)
	}
	if pkgs[1].Name != "foo" {
		t.Errorf("second package: got %q, want foo", pkgs[1].Name)
	}
}
