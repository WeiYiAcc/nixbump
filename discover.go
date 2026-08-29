package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type Package struct {
	Name   string
	Dir    string
	Config *Config
}

// configFileNames: yaml is the source of truth (comment-friendly, sops-native);
// json kept for backward compatibility with upstream nixbump.
var configFileNames = []string{"nixbump.yaml", "nixbump.yml", "nixbump.json"}

func discoverPackages(pkgsDir string) ([]Package, error) {
	entries, err := os.ReadDir(pkgsDir)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", pkgsDir, err)
	}

	var packages []Package
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		var cfgPath string
		for _, name := range configFileNames {
			p := filepath.Join(pkgsDir, e.Name(), name)
			if _, err := os.Stat(p); err == nil {
				cfgPath = p
				break
			}
		}
		if cfgPath == "" {
			continue
		}
		cfg, err := parseConfig(cfgPath)
		if err != nil {
			return nil, fmt.Errorf("package %s: %w", e.Name(), err)
		}
		packages = append(packages, Package{
			Name:   e.Name(),
			Dir:    filepath.Join(pkgsDir, e.Name()),
			Config: cfg,
		})
	}

	sort.Slice(packages, func(i, j int) bool {
		return packages[i].Name < packages[j].Name
	})

	return packages, nil
}
