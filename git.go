package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func gitRevParseTopLevel() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not inside a git repository")
	}
	return strings.TrimSpace(string(out)), nil
}

func gitHasChanges(dir string) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) != "", nil
}

func gitWorktreeAdd(repoDir, worktreePath, branch string) error {
	cmd := exec.Command("git", "worktree", "add", "-b", branch, worktreePath)
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git worktree add: %s", string(out))
	}
	return nil
}

func gitWorktreeRemove(repoDir, worktreePath string) error {
	cmd := exec.Command("git", "worktree", "remove", "--force", worktreePath)
	cmd.Dir = repoDir
	cmd.CombinedOutput()
	return nil
}

func gitBranchDelete(repoDir, branch string) {
	cmd := exec.Command("git", "branch", "-D", branch)
	cmd.Dir = repoDir
	cmd.CombinedOutput()
}

func gitBranchExistsRemote(repoDir, branch string) bool {
	cmd := exec.Command("git", "ls-remote", "--heads", "origin", branch)
	cmd.Dir = repoDir
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}

func gitAddAll(dir string) error {
	cmd := exec.Command("git", "add", "-A")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add: %s", string(out))
	}
	return nil
}

func gitCommit(dir, message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit: %s", string(out))
	}
	return nil
}

func gitPush(dir, branch string) error {
	cmd := exec.Command("git", "push", "-u", "origin", branch)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git push: %s", string(out))
	}
	return nil
}

func ghPRCreate(dir, title, body string) (string, error) {
	cmd := exec.Command("gh", "pr", "create", "--title", title, "--body", body)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("gh pr create: %s", string(out))
	}
	return strings.TrimSpace(string(out)), nil
}
