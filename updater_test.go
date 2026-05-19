package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseNixVersion(t *testing.T) {
	content := `  version = "1.6.6";`
	got := parseNixVersion(content)
	if got != "1.6.6" {
		t.Errorf("got %q, want 1.6.6", got)
	}
}

func TestReplaceVersion(t *testing.T) {
	content := `  pname = "foo";
  version = "1.0.0";
`
	got := replaceVersion(content, "1.0.0", "2.0.0")
	if !strings.Contains(got, `version = "2.0.0"`) {
		t.Errorf("version not replaced:\n%s", got)
	}
	if strings.Contains(got, `version = "1.0.0"`) {
		t.Errorf("old version still present:\n%s", got)
	}
}

func TestReplaceHash(t *testing.T) {
	content := `  src = fetchurl {
    url = "https://example.com/${version}/foo.tar.gz";
    hash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
  };`
	got := replaceHash(content, "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=", "sha256-BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB=")
	if !strings.Contains(got, "sha256-BBB") {
		t.Errorf("hash not replaced:\n%s", got)
	}
}

func TestParseNixHashes(t *testing.T) {
	data, err := os.ReadFile("testdata/multi-src/default.nix")
	if err != nil {
		t.Fatal(err)
	}
	hashes := parseNixHashes(string(data))
	if len(hashes) != 2 {
		t.Fatalf("got %d hashes, want 2", len(hashes))
	}
}

func TestUpdateNixFile(t *testing.T) {
	dir := t.TempDir()
	data, err := os.ReadFile("testdata/simple/default.nix")
	if err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dir, "default.nix")
	os.WriteFile(dst, data, 0644)

	err = updateNixVersion(dst, "0.2.20", "0.3.0")
	if err != nil {
		t.Fatal(err)
	}

	updated, _ := os.ReadFile(dst)
	if !strings.Contains(string(updated), `version = "0.3.0"`) {
		t.Errorf("version not updated:\n%s", string(updated))
	}
}
