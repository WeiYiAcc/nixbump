package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

type releaseInfo struct {
	Version string
	Tag     string
}

func extractVersion(tag string, cfg *Config) (string, error) {
	if cfg.VersionRegex != "" {
		re, err := regexp.Compile(cfg.VersionRegex)
		if err != nil {
			return "", fmt.Errorf("invalid version_regex: %w", err)
		}
		m := re.FindStringSubmatch(tag)
		if len(m) < 2 {
			return "", fmt.Errorf("version_regex %q did not match tag %q", cfg.VersionRegex, tag)
		}
		return m[1], nil
	}
	if cfg.StripPrefix != "" {
		return strings.TrimPrefix(tag, cfg.StripPrefix), nil
	}
	return tag, nil
}

func fetchLatestVersion(cfg *Config) (*releaseInfo, error) {
	switch cfg.Source {
	case "github":
		return fetchGitHubRelease(cfg)
	case "gitlab":
		return fetchGitLabRelease(cfg)
	case "npm":
		return fetchNpmVersion(cfg)
	default:
		return nil, fmt.Errorf("unsupported source for version fetch: %s", cfg.Source)
	}
}

func fetchGitHubRelease(cfg *Config) (*releaseInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", cfg.Owner, cfg.Repo)
	var data struct {
		TagName string `json:"tag_name"`
	}
	if err := fetchJSON(url, &data); err != nil {
		return nil, fmt.Errorf("github %s/%s: %w", cfg.Owner, cfg.Repo, err)
	}
	ver, err := extractVersion(data.TagName, cfg)
	if err != nil {
		return nil, err
	}
	return &releaseInfo{Version: ver, Tag: data.TagName}, nil
}

func fetchGitLabRelease(cfg *Config) (*releaseInfo, error) {
	host := cfg.Host
	if host == "" {
		host = "gitlab.com"
	}
	project := fmt.Sprintf("%s%%2F%s", cfg.Owner, cfg.Repo)
	url := fmt.Sprintf("https://%s/api/v4/projects/%s/releases", host, project)

	var releases []struct {
		TagName string `json:"tag_name"`
	}
	if err := fetchJSON(url, &releases); err != nil {
		return nil, fmt.Errorf("gitlab %s/%s: %w", cfg.Owner, cfg.Repo, err)
	}
	if len(releases) == 0 {
		return nil, fmt.Errorf("gitlab %s/%s: no releases found", cfg.Owner, cfg.Repo)
	}
	ver, err := extractVersion(releases[0].TagName, cfg)
	if err != nil {
		return nil, err
	}
	return &releaseInfo{Version: ver, Tag: releases[0].TagName}, nil
}

func fetchNpmVersion(cfg *Config) (*releaseInfo, error) {
	url := fmt.Sprintf("https://registry.npmjs.org/%s/latest", cfg.Package)
	var data struct {
		Version string `json:"version"`
	}
	if err := fetchJSON(url, &data); err != nil {
		return nil, fmt.Errorf("npm %s: %w", cfg.Package, err)
	}
	return &releaseInfo{Version: data.Version, Tag: data.Version}, nil
}

func fetchJSON(url string, target any) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}
	return json.NewDecoder(resp.Body).Decode(target)
}
