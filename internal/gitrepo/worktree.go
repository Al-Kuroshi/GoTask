package gitrepo

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
)

// Repo wraps git CLI operations against the local clone at Path, authenticating
// pushes with Token (a GitHub PAT) via an injected HTTP header rather than
// mutating global git config or credential helpers.
type Repo struct {
	Path  string
	Token string
}

func New(path, token string) *Repo {
	return &Repo{Path: path, Token: token}
}

func (r *Repo) gitArgs(args ...string) []string {
	base := []string{
		"-c", "user.name=GoTask Bot",
		"-c", "user.email=gotask-bot@localhost",
	}
	if r.Token != "" {
		header := "AUTHORIZATION: basic " + base64.StdEncoding.EncodeToString([]byte("x-access-token:"+r.Token))
		base = append(base, "-c", "http.extraheader="+header)
	}
	return append(base, args...)
}

func (r *Repo) runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", r.gitArgs(args...)...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("git %v failed: %w: %s", args, err, string(out))
	}
	return string(out), nil
}

// NewOrphanWorktree creates an isolated worktree checked out to a brand-new
// orphan branch (no shared history with main) so a person's branch never
// carries this service's own source code. Returns the worktree path and a
// cleanup function that must be called once done with it.
func (r *Repo) NewOrphanWorktree(branchName string) (string, func(), error) {
	tmpDir, err := os.MkdirTemp("", "gotask-wt-*")
	if err != nil {
		return "", nil, fmt.Errorf("create temp dir: %w", err)
	}
	cleanup := func() {
		r.runGit(r.Path, "worktree", "remove", "--force", tmpDir)
		os.RemoveAll(tmpDir)
	}

	if _, err := r.runGit(r.Path, "worktree", "add", "--detach", tmpDir, "HEAD"); err != nil {
		os.RemoveAll(tmpDir)
		return "", nil, err
	}
	if _, err := r.runGit(tmpDir, "checkout", "--orphan", branchName); err != nil {
		cleanup()
		return "", nil, err
	}
	// checkout --orphan keeps the previous branch's tracked files staged; clear
	// them so the new branch starts genuinely empty.
	r.runGit(tmpDir, "rm", "-rf", "--ignore-unmatch", ".")

	return tmpDir, cleanup, nil
}

// AddWorktree checks out an already-existing branch into an isolated worktree.
func (r *Repo) AddWorktree(branchName string) (string, func(), error) {
	tmpDir, err := os.MkdirTemp("", "gotask-wt-*")
	if err != nil {
		return "", nil, fmt.Errorf("create temp dir: %w", err)
	}
	cleanup := func() {
		r.runGit(r.Path, "worktree", "remove", "--force", tmpDir)
		os.RemoveAll(tmpDir)
	}

	r.runGit(r.Path, "fetch", "origin", branchName)
	if _, err := r.runGit(r.Path, "worktree", "add", tmpDir, branchName); err != nil {
		os.RemoveAll(tmpDir)
		return "", nil, err
	}
	return tmpDir, cleanup, nil
}
