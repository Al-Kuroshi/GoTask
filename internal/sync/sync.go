package sync

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/Al-Kuroshi/GoTask/internal/clickup"
	"github.com/Al-Kuroshi/GoTask/internal/db"
	"github.com/Al-Kuroshi/GoTask/internal/gitrepo"
)

// SyncAssignee fetches the assignee's current tasks from ClickUp and writes
// one commit per task into their branch. Per-task failures are logged and
// skipped rather than aborting the whole run; the overall sync_status reflects
// whether anything failed.
func SyncAssignee(database *sql.DB, repo *gitrepo.Repo, cu *clickup.Client, teamID string, assigneeID int64) error {
	a, err := db.GetAssignee(database, assigneeID)
	if err != nil {
		return err
	}

	if err := db.UpdateSyncStatus(database, a.ID, db.StatusSyncing, ""); err != nil {
		return err
	}

	tasks, err := cu.GetTeamTasksByAssignee(teamID, a.ClickUpUserID)
	if err != nil {
		db.UpdateSyncStatus(database, a.ID, db.StatusError, err.Error())
		return fmt.Errorf("fetch tasks: %w", err)
	}

	wt, cleanup, err := repo.AddWorktree(a.BranchName)
	if err != nil {
		db.UpdateSyncStatus(database, a.ID, db.StatusError, err.Error())
		return fmt.Errorf("checkout branch: %w", err)
	}
	defer cleanup()

	var syncErrs []string
	for _, t := range tasks {
		comments, err := cu.GetTaskComments(t.ID)
		if err != nil {
			syncErrs = append(syncErrs, fmt.Sprintf("task %s comments: %v", t.ID, err))
			continue
		}
		if err := gitrepo.WriteTaskDir(wt, t, comments, a.ClickUpUsername, cu.DownloadAttachment); err != nil {
			syncErrs = append(syncErrs, fmt.Sprintf("task %s: %v", t.ID, err))
			continue
		}
		if err := repo.CommitAll(wt, fmt.Sprintf("Sync task %s: %s", t.ID, t.Name)); err != nil {
			syncErrs = append(syncErrs, fmt.Sprintf("task %s commit: %v", t.ID, err))
		}
	}

	if err := repo.Push(wt, a.BranchName); err != nil {
		syncErrs = append(syncErrs, fmt.Sprintf("push: %v", err))
	}

	if len(syncErrs) > 0 {
		joined := strings.Join(syncErrs, "; ")
		db.UpdateSyncStatus(database, a.ID, db.StatusError, joined)
		return fmt.Errorf("sync completed with errors: %s", joined)
	}

	return db.MarkSynced(database, a.ID)
}
