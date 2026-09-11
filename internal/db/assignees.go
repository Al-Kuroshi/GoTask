package db

import (
	"database/sql"
	"fmt"
	"time"
)

type SyncStatus string

const (
	StatusPending SyncStatus = "pending"
	StatusSyncing SyncStatus = "syncing"
	StatusSynced  SyncStatus = "synced"
	StatusError   SyncStatus = "error"
)

type Assignee struct {
	ID              int64
	ClickUpUserID   string
	ClickUpUsername string
	GitHubUsername  string
	WhatsAppNumber  string
	BranchName      string
	SyncStatus      SyncStatus
	SyncError       string
	LastSyncedAt    sql.NullTime
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func InsertAssignee(database *sql.DB, a *Assignee) (int64, error) {
	res, err := database.Exec(
		`INSERT INTO assignees
			(clickup_user_id, clickup_username, github_username, whatsapp_number, branch_name, sync_status)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		a.ClickUpUserID, a.ClickUpUsername, nullIfEmpty(a.GitHubUsername), nullIfEmpty(a.WhatsAppNumber),
		a.BranchName, string(StatusPending),
	)
	if err != nil {
		return 0, fmt.Errorf("insert assignee: %w", err)
	}
	return res.LastInsertId()
}

func GetAssignee(database *sql.DB, id int64) (*Assignee, error) {
	return scanAssignee(database.QueryRow(
		`SELECT id, clickup_user_id, clickup_username, github_username, whatsapp_number,
		        branch_name, sync_status, sync_error, last_synced_at, created_at, updated_at
		 FROM assignees WHERE id = ?`, id))
}

func GetAssigneeByClickUpID(database *sql.DB, clickupUserID string) (*Assignee, error) {
	return scanAssignee(database.QueryRow(
		`SELECT id, clickup_user_id, clickup_username, github_username, whatsapp_number,
		        branch_name, sync_status, sync_error, last_synced_at, created_at, updated_at
		 FROM assignees WHERE clickup_user_id = ?`, clickupUserID))
}

func scanAssignee(row *sql.Row) (*Assignee, error) {
	var a Assignee
	var githubUsername, whatsappNumber, syncError sql.NullString
	var status string
	if err := row.Scan(
		&a.ID, &a.ClickUpUserID, &a.ClickUpUsername, &githubUsername, &whatsappNumber,
		&a.BranchName, &status, &syncError, &a.LastSyncedAt, &a.CreatedAt, &a.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("scan assignee: %w", err)
	}
	a.GitHubUsername = githubUsername.String
	a.WhatsAppNumber = whatsappNumber.String
	a.SyncError = syncError.String
	a.SyncStatus = SyncStatus(status)
	return &a, nil
}

func UpdateSyncStatus(database *sql.DB, id int64, status SyncStatus, syncErr string) error {
	_, err := database.Exec(
		`UPDATE assignees SET sync_status = ?, sync_error = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		string(status), nullIfEmpty(syncErr), id,
	)
	if err != nil {
		return fmt.Errorf("update sync status: %w", err)
	}
	return nil
}

func MarkSynced(database *sql.DB, id int64) error {
	_, err := database.Exec(
		`UPDATE assignees
		 SET sync_status = ?, sync_error = NULL, last_synced_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		string(StatusSynced), id,
	)
	if err != nil {
		return fmt.Errorf("mark synced: %w", err)
	}
	return nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
