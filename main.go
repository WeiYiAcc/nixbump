package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type updateResult struct {
	pkg        Package
	oldVersion string
	newVersion string
	changed    bool
	err        error
}

func main() {
	dryRun := flag.Bool("dry-run", false, "show what would be done without making changes")
	pkgName := flag.String("package", "", "update only the specified package")
	pkgShort := flag.String("p", "", "update only the specified package (shorthand)")
	list := flag.Bool("list", false, "list all discoverable packages")
	listShort := flag.Bool("l", false, "list all discoverable packages (shorthand)")
	pr := flag.Bool("pr", false, "create a PR for each updated package")
	flag.Parse()

	if *pkgShort != "" && *pkgName == "" {
		*pkgName = *pkgShort
	}
	if *listShort {
		*list = true
	}

	repoRoot, err := gitRevParseTopLevel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	pkgsDir := filepath.Join(repoRoot, "pkgs")
	if _, err := os.Stat(pkgsDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error: no pkgs/ directory found in %s\n", repoRoot)
		os.Exit(1)
	}

	packages, err := discoverPackages(pkgsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *list {
		printPackageList(packages)
		return
	}

	if *pkgName != "" {
		filtered := filterPackages(packages, *pkgName)
		if len(filtered) == 0 {
			fmt.Fprintf(os.Stderr, "Error: package %q not found\n", *pkgName)
			os.Exit(1)
		}
		packages = filtered
	}

	var failures int
	for _, pkg := range packages {
		var result updateResult
		if *pr {
			result = updateWithPR(pkg, repoRoot, *dryRun)
		} else {
			result = updateInPlace(pkg, repoRoot, *dryRun)
		}

		if result.err != nil {
			fmt.Fprintf(os.Stderr, "  FAIL: %v\n", result.err)
			failures++
		} else if result.changed {
			fmt.Printf("  %s -> %s\n", result.oldVersion, result.newVersion)
		} else {
			fmt.Printf("  up to date (%s)\n", result.oldVersion)
		}
	}

	if failures > 0 {
		os.Exit(1)
	}
}

func printPackageList(packages []Package) {
	for _, pkg := range packages {
		nixPath := filepath.Join(pkg.Dir, "default.nix")
		data, _ := os.ReadFile(nixPath)
		ver := parseNixVersion(string(data))
		fmt.Printf("  %-30s %s (%s)\n", pkg.Name, ver, pkg.Config.Source)
	}
}

func filterPackages(packages []Package, name string) []Package {
	var filtered []Package
	for _, pkg := range packages {
		if pkg.Name == name {
			filtered = append(filtered, pkg)
		}
	}
	return filtered
}

func updateInPlace(pkg Package, repoRoot string, dryRun bool) updateResult {
	fmt.Printf("\n%s (%s)...\n", pkg.Name, pkg.Config.Source)

	nixPath := filepath.Join(pkg.Dir, "default.nix")
	data, err := os.ReadFile(nixPath)
	if err != nil {
		return updateResult{pkg: pkg, err: err}
	}

	oldVersion := parseNixVersion(string(data))

	if pkg.Config.Source == "nix-update" {
		if dryRun {
			fmt.Println("  (dry-run, would run nix-update)")
			return updateResult{pkg: pkg, oldVersion: oldVersion}
		}
		err := runNixUpdate(pkg, repoRoot)
		newData, _ := os.ReadFile(nixPath)
		newVersion := parseNixVersion(string(newData))
		return updateResult{
			pkg: pkg, oldVersion: oldVersion, newVersion: newVersion,
			changed: oldVersion != newVersion, err: err,
		}
	}

	info, err := fetchLatestVersion(pkg.Config)
	if err != nil {
		return updateResult{pkg: pkg, oldVersion: oldVersion, err: err}
	}

	if info.Version == oldVersion {
		return updateResult{pkg: pkg, oldVersion: oldVersion, newVersion: oldVersion}
	}

	if dryRun {
		fmt.Printf("  would update: %s -> %s\n", oldVersion, info.Version)
		return updateResult{pkg: pkg, oldVersion: oldVersion, newVersion: info.Version, changed: true}
	}

	if err := updateNixVersion(nixPath, oldVersion, info.Version); err != nil {
		return updateResult{pkg: pkg, oldVersion: oldVersion, err: err}
	}

	updatedData, _ := os.ReadFile(nixPath)
	oldHashes := parseNixHashes(string(data))
	if len(oldHashes) == 0 {
		return updateResult{pkg: pkg, oldVersion: oldVersion, newVersion: info.Version, changed: true}
	}

	newHashes, err := prefetchNewHashes(pkg, info, string(updatedData))
	if err != nil {
		updateNixVersion(nixPath, info.Version, oldVersion)
		return updateResult{pkg: pkg, oldVersion: oldVersion, err: fmt.Errorf("hash prefetch: %w", err)}
	}

	for i, oldHash := range oldHashes {
		if i < len(newHashes) {
			if err := updateNixHash(nixPath, oldHash, newHashes[i]); err != nil {
				return updateResult{pkg: pkg, oldVersion: oldVersion, err: err}
			}
		}
	}

	return updateResult{pkg: pkg, oldVersion: oldVersion, newVersion: info.Version, changed: true}
}

func prefetchNewHashes(pkg Package, info *releaseInfo, nixContent string) ([]string, error) {
	urlRe := regexp.MustCompile(`url\s*=\s*"([^"]+)"`)
	matches := urlRe.FindAllStringSubmatch(nixContent, -1)

	var hashes []string
	for _, m := range matches {
		url := m[1]
		url = strings.ReplaceAll(url, "${version}", info.Version)

		fmt.Printf("  prefetching: %s\n", url)

		unpack := strings.HasSuffix(url, ".tar.gz") ||
			strings.HasSuffix(url, ".tar.bz2") ||
			strings.HasSuffix(url, ".tar.xz") ||
			strings.HasSuffix(url, ".zip")

		hash, err := nixPrefetchURL(url, unpack)
		if err != nil {
			return nil, fmt.Errorf("prefetch %s: %w", url, err)
		}
		hashes = append(hashes, hash)
	}

	return hashes, nil
}

func runNixUpdate(pkg Package, repoRoot string) error {
	args := []string{"--flake"}
	args = append(args, pkg.Config.Args...)
	cmd := exec.Command("nix-update", args...)
	cmd.Dir = repoRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func updateWithPR(pkg Package, repoRoot string, dryRun bool) updateResult {
	fmt.Printf("\n%s (%s) [PR mode]...\n", pkg.Name, pkg.Config.Source)

	nixPath := filepath.Join(pkg.Dir, "default.nix")
	data, _ := os.ReadFile(nixPath)
	oldVersion := parseNixVersion(string(data))

	branch := fmt.Sprintf("update/%s", pkg.Name)

	if gitBranchExistsRemote(repoRoot, branch) {
		fmt.Printf("  branch %s already exists on remote, skipping\n", branch)
		return updateResult{pkg: pkg, oldVersion: oldVersion}
	}

	if dryRun {
		fmt.Println("  (dry-run, would create worktree + PR)")
		return updateResult{pkg: pkg, oldVersion: oldVersion}
	}

	tmpDir, err := os.MkdirTemp("", "nixbump-*")
	if err != nil {
		return updateResult{pkg: pkg, oldVersion: oldVersion, err: err}
	}
	defer os.RemoveAll(tmpDir)

	worktreePath := filepath.Join(tmpDir, "worktree")

	if err := gitWorktreeAdd(repoRoot, worktreePath, branch); err != nil {
		return updateResult{pkg: pkg, oldVersion: oldVersion, err: err}
	}
	defer func() {
		gitWorktreeRemove(repoRoot, worktreePath)
		gitBranchDelete(repoRoot, branch)
	}()

	wtPkg := Package{
		Name:   pkg.Name,
		Dir:    filepath.Join(worktreePath, "pkgs", pkg.Name),
		Config: pkg.Config,
	}
	result := updateInPlace(wtPkg, worktreePath, false)
	if result.err != nil || !result.changed {
		return result
	}

	if err := gitAddAll(worktreePath); err != nil {
		return updateResult{pkg: pkg, oldVersion: oldVersion, err: err}
	}

	commitMsg := fmt.Sprintf("%s: %s -> %s", pkg.Name, result.oldVersion, result.newVersion)
	if err := gitCommit(worktreePath, commitMsg); err != nil {
		return updateResult{pkg: pkg, oldVersion: oldVersion, err: err}
	}

	if err := gitPush(worktreePath, branch); err != nil {
		return updateResult{pkg: pkg, oldVersion: oldVersion, err: err}
	}

	prBody := fmt.Sprintf("Automated update of %s from %s to %s.", pkg.Name, result.oldVersion, result.newVersion)
	prURL, err := ghPRCreate(worktreePath, commitMsg, prBody)
	if err != nil {
		return updateResult{pkg: pkg, oldVersion: oldVersion, err: err}
	}
	fmt.Printf("  PR: %s\n", prURL)

	return result
}
