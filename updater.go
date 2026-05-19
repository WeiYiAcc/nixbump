package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	versionRe = regexp.MustCompile(`version\s*=\s*"([^"]+)"`)
	hashRe    = regexp.MustCompile(`(sha256-[A-Za-z0-9+/]+=*)`)
)

func parseNixVersion(content string) string {
	m := versionRe.FindStringSubmatch(content)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

func parseNixHashes(content string) []string {
	return hashRe.FindAllString(content, -1)
}

func replaceVersion(content, oldVer, newVer string) string {
	old := fmt.Sprintf(`version = "%s"`, oldVer)
	new := fmt.Sprintf(`version = "%s"`, newVer)
	return strings.Replace(content, old, new, 1)
}

func replaceHash(content, oldHash, newHash string) string {
	return strings.Replace(content, oldHash, newHash, 1)
}

func updateNixVersion(path, oldVer, newVer string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := replaceVersion(string(data), oldVer, newVer)
	return os.WriteFile(path, []byte(content), 0644)
}

func updateNixHash(path, oldHash, newHash string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := replaceHash(string(data), oldHash, newHash)
	return os.WriteFile(path, []byte(content), 0644)
}
