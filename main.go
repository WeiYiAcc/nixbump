package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

type updateResult struct {
	pkg        Package
	oldVersion string
	newVersion string
	changed    bool
	err        error
}

// mtpSchema is the --mtp-describe payload (HN nr378: JSON describing
// commands, args, types, examples; no server, no transport, no handshake).
type mtpSchema struct {
	Version  string       `json:"version"`
	Name     string       `json:"name"`
	Commands []mtpCommand `json:"commands"`
}

type mtpCommand struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Args        []string `json:"args,omitempty"`
	Examples    []string `json:"examples,omitempty"`
}

func mtpDescribe() mtpSchema {
	return mtpSchema{
		Version: "1.0",
		Name:    "nixbump",
		Commands: []mtpCommand{
			{Name: "list", Description: "list all discoverable packages"},
			{Name: "check", Description: "dry-run: show available updates without changing files", Examples: []string{"nixbump check", "nixbump check atomic"}},
			{Name: "update", Description: "update one package (or all) in-place", Args: []string{"[pkg]"}, Examples: []string{"nixbump update atomic"}},
			{Name: "pr", Description: "update each outdated package and create a PR via git worktree + gh"},
		},
	}
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "nixbump",
		Short: "Auto-discover and update custom Nix packages (npm/GitHub/GitLab)",
		Long: `nixbump discovers pkgs/*/nixbump.yaml configs, fetches the latest
version from the configured source, and rewrites version + sha256 in the
package derivation in-place. Configs may be sops-encrypted YAML.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.CompletionOptions.HiddenDefaultCmd = true
	return root
}

func requireRepoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("not inside a git repo: %w", err)
	}
	root := filepath.Clean(strings.TrimSpace(string(out)))
	pkgsDir := filepath.Join(root, "pkgs")
	if _, err := os.Stat(pkgsDir); err != nil {
		return "", fmt.Errorf("pkgs/ not found under %s", root)
	}
	return root, nil
}

func loadPackages() ([]Package, error) {
	root, err := requireRepoRoot()
	if err != nil {
		return nil, err
	}
	return discoverPackages(filepath.Join(root, "pkgs"))
}

func runList() error {
	pkgs, err := loadPackages()
	if err != nil {
		return err
	}
	printPackageList(pkgs)
	return nil
}

func runUpdate(pkgName string, dryRun, createPR bool) error {
	pkgs, err := loadPackages()
	if err != nil {
		return err
	}
	if pkgName != "" {
		pkgs = filterPackages(pkgs, pkgName)
		if len(pkgs) == 0 {
			return fmt.Errorf("package not found: %s", pkgName)
		}
	}
	root, _ := requireRepoRoot()

	var failures int
	for _, pkg := range pkgs {
		var res updateResult
		if createPR {
			res = updateWithPR(pkg, root, dryRun)
		} else {
			res = updateInPlace(pkg, root, dryRun)
		}
		if res.err != nil {
			fmt.Printf("  FAIL: %v\n", res.err)
			failures++
		}
	}
	if failures > 0 {
		os.Exit(1)
	}
	return nil
}

func main() {
	root := newRootCmd()

	// AI discovery (MTP): nixbump --mtp-describe
	root.Flags().Bool("mtp-describe", false, "print MTP JSON schema for AI discovery and exit")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "list all discoverable packages",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList()
		},
	}

	checkCmd := &cobra.Command{
		Use:   "check [pkg]",
		Short: "dry-run: show available updates without changing files",
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			return runUpdate(name, true, false)
		},
	}

	updateCmd := &cobra.Command{
		Use:   "update [pkg]",
		Short: "update one package (or all) in-place",
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			return runUpdate(name, false, false)
		},
	}

	prCmd := &cobra.Command{
		Use:   "pr [pkg]",
		Short: "update and create a PR per package (worktree + gh)",
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			return runUpdate(name, false, true)
		},
	}

	root.AddCommand(listCmd, checkCmd, updateCmd, prCmd)

	describe := false
	for _, a := range os.Args[1:] {
		if a == "--mtp-describe" {
			describe = true
		}
	}
	if describe {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(mtpDescribe())
		return
	}

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
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
