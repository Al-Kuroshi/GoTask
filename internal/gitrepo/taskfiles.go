package gitrepo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Al-Kuroshi/GoTask/internal/clickup"
	"gopkg.in/yaml.v3"
)

type Provenance struct {
	ReassignedFrom string `yaml:"reassigned_from,omitempty"`
	SourceCommit   string `yaml:"source_commit,omitempty"`
	ReassignedAt   string `yaml:"reassigned_at,omitempty"`
}

type TaskYAML struct {
	ID          string      `yaml:"id"`
	Title       string      `yaml:"title"`
	Status      string      `yaml:"status"`
	Priority    string      `yaml:"priority,omitempty"`
	DueDate     string      `yaml:"due_date,omitempty"`
	Assignee    string      `yaml:"assignee"`
	CreatedAt   string      `yaml:"created_at,omitempty"`
	UpdatedAt   string      `yaml:"updated_at,omitempty"`
	Provenance  *Provenance `yaml:"provenance,omitempty"`
	Attachments []string    `yaml:"attachments,omitempty"`
}

// DownloadFunc fetches the raw bytes of an attachment URL.
type DownloadFunc func(url string) ([]byte, error)

// WriteTaskDir renders one task's directory (task.yaml, description.md,
// checklist.yaml, comments.md, attachments/) into the given worktree.
func WriteTaskDir(worktreePath string, t clickup.Task, comments []clickup.Comment, assigneeUsername string, download DownloadFunc) error {
	dir := filepath.Join(worktreePath, "task-"+t.ID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create task dir: %w", err)
	}

	attachmentNames, err := writeAttachments(dir, t.Attachments, download)
	if err != nil {
		return err
	}

	taskYAML := TaskYAML{
		ID:          t.ID,
		Title:       t.Name,
		Status:      t.Status.Status,
		Priority:    t.Priority.String(),
		DueDate:     t.DueDate,
		Assignee:    assigneeUsername,
		CreatedAt:   t.DateCreated,
		UpdatedAt:   t.DateUpdated,
		Attachments: attachmentNames,
	}
	yamlBytes, err := yaml.Marshal(taskYAML)
	if err != nil {
		return fmt.Errorf("marshal task.yaml: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "task.yaml"), yamlBytes, 0o644); err != nil {
		return fmt.Errorf("write task.yaml: %w", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "description.md"), []byte(t.Description), 0o644); err != nil {
		return fmt.Errorf("write description.md: %w", err)
	}

	checklistBytes, err := yaml.Marshal(t.Checklists)
	if err != nil {
		return fmt.Errorf("marshal checklist.yaml: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "checklist.yaml"), checklistBytes, 0o644); err != nil {
		return fmt.Errorf("write checklist.yaml: %w", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "comments.md"), []byte(renderComments(comments)), 0o644); err != nil {
		return fmt.Errorf("write comments.md: %w", err)
	}

	return nil
}

func writeAttachments(dir string, attachments []clickup.Attachment, download DownloadFunc) ([]string, error) {
	if len(attachments) == 0 {
		return nil, nil
	}
	attachDir := filepath.Join(dir, "attachments")
	if err := os.MkdirAll(attachDir, 0o755); err != nil {
		return nil, fmt.Errorf("create attachments dir: %w", err)
	}

	var names []string
	for _, a := range attachments {
		data, err := download(a.URL)
		if err != nil {
			return names, fmt.Errorf("download attachment %s: %w", a.Title, err)
		}
		name := a.Title
		if name == "" {
			name = a.ID
		}
		if err := os.WriteFile(filepath.Join(attachDir, name), data, 0o644); err != nil {
			return names, fmt.Errorf("write attachment %s: %w", name, err)
		}
		names = append(names, name)
	}
	return names, nil
}

func renderComments(comments []clickup.Comment) string {
	var b strings.Builder
	for _, c := range comments {
		fmt.Fprintf(&b, "## %s — %s\n\n%s\n\n", c.User.Username, c.Date, c.CommentText)
	}
	return b.String()
}
