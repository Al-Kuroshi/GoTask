package sync

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Al-Kuroshi/GoTask/internal/db"
	"github.com/Al-Kuroshi/GoTask/internal/gitrepo"
)

// ComputeBranchName applies the <clickup_username>-<github_username> convention,
// falling back to just the ClickUp username when no GitHub account is on file.
func ComputeBranchName(clickupUsername, githubUsername string) string {
	if githubUsername == "" {
		return clickupUsername
	}
	return clickupUsername + "-" + githubUsername
}

type RegisterParams struct {
	ClickUpUserID   string
	ClickUpUsername string
	GitHubUsername  string
	WhatsAppNumber  string
}

// RegisterAssignee records a new assignee and creates their orphan branch on
// GitHub. It does not perform the ClickUp sync itself — call SyncAssignee with
// the returned ID afterwards.
func RegisterAssignee(database *sql.DB, repo *gitrepo.Repo, p RegisterParams) (int64, error) {
	branch := ComputeBranchName(p.ClickUpUsername, p.GitHubUsername)

	id, err := db.InsertAssignee(database, &db.Assignee{
		ClickUpUserID:   p.ClickUpUserID,
		ClickUpUsername: p.ClickUpUsername,
		GitHubUsername:  p.GitHubUsername,
		WhatsAppNumber:  p.WhatsAppNumber,
		BranchName:      branch,
	})
	if err != nil {
		return 0, fmt.Errorf("insert assignee: %w", err)
	}

	wt, cleanup, err := repo.NewOrphanWorktree(branch)
	if err != nil {
		db.UpdateSyncStatus(database, id, db.StatusError, err.Error())
		return id, fmt.Errorf("create branch: %w", err)
	}
	defer cleanup()

	if err := os.WriteFile(filepath.Join(wt, ".gitkeep"), []byte{}, 0o644); err != nil {
		db.UpdateSyncStatus(database, id, db.StatusError, err.Error())
		return id, fmt.Errorf("write .gitkeep: %w", err)
	}
	if err := repo.CommitAll(wt, "Initialize branch for "+p.ClickUpUsername); err != nil {
		db.UpdateSyncStatus(database, id, db.StatusError, err.Error())
		return id, fmt.Errorf("initial commit: %w", err)
	}
	if err := repo.Push(wt, branch); err != nil {
		db.UpdateSyncStatus(database, id, db.StatusError, err.Error())
		return id, fmt.Errorf("push branch: %w", err)
	}

	return id, nil
}
