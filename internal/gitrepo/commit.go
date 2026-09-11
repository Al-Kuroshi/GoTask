package gitrepo

// CommitAll stages every change in the worktree and commits it. Callers are
// expected to make one commit per logical change (one task, one field update)
// so branch history stays a readable ledger of real events.
func (r *Repo) CommitAll(worktreePath, message string) error {
	if _, err := r.runGit(worktreePath, "add", "-A"); err != nil {
		return err
	}
	if _, err := r.runGit(worktreePath, "commit", "-m", message); err != nil {
		return err
	}
	return nil
}

func (r *Repo) Push(worktreePath, branchName string) error {
	_, err := r.runGit(worktreePath, "push", "-u", "origin", branchName)
	return err
}
