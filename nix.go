package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func nixPrefetchURL(url string, unpack bool) (string, error) {
	args := []string{url}
	if unpack {
		args = append([]string{"--unpack"}, args...)
	}

	cmd := exec.Command("nix-prefetch-url", args...)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("nix-prefetch-url: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("nix-prefetch-url: %w", err)
	}

	base32Hash := strings.TrimSpace(string(out))
	return nixHashToSRI(base32Hash)
}

func nixHashToSRI(base32Hash string) (string, error) {
	cmd := exec.Command("nix", "hash", "to-sri", "--type", "sha256", base32Hash)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("nix hash to-sri: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func prefetchGitHubArchive(owner, repo, rev string) (string, error) {
	url := fmt.Sprintf("https://github.com/%s/%s/archive/%s.tar.gz", owner, repo, rev)
	return nixPrefetchURL(url, true)
}

func prefetchGitLabArchive(host, owner, repo, rev string) (string, error) {
	url := fmt.Sprintf("https://%s/%s/%s/-/archive/%s/%s-%s.tar.gz", host, owner, repo, rev, repo, rev)
	return nixPrefetchURL(url, true)
}
