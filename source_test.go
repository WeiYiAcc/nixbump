package main

import (
	"testing"
)

func TestExtractVersion(t *testing.T) {
	tests := []struct {
		tag          string
		stripPrefix  string
		versionRegex string
		want         string
	}{
		{"v1.2.3", "v", "", "1.2.3"},
		{"v1.2.3", "", "", "v1.2.3"},
		{"release-1.2.3", "", `release-(.*)`, "1.2.3"},
		{"server/v2.1.2", "", `server/v(.*)`, "2.1.2"},
	}

	for _, tt := range tests {
		cfg := &Config{StripPrefix: tt.stripPrefix, VersionRegex: tt.versionRegex}
		got, err := extractVersion(tt.tag, cfg)
		if err != nil {
			t.Errorf("extractVersion(%q): %v", tt.tag, err)
			continue
		}
		if got != tt.want {
			t.Errorf("extractVersion(%q) = %q, want %q", tt.tag, got, tt.want)
		}
	}
}
