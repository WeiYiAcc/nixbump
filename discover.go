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
		cfgPath := filepath.Join(pkgsDir, e.Name(), "nixbump.json")
		if _, err := os.Stat(cfgPath); err != nil {
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
