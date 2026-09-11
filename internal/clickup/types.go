package clickup

type Task struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Status      Status       `json:"status"`
	Priority    *Priority    `json:"priority"`
	DueDate     string       `json:"due_date"`
	DateCreated string       `json:"date_created"`
	DateUpdated string       `json:"date_updated"`
	Checklists  []Checklist  `json:"checklists"`
	Attachments []Attachment `json:"attachments"`
	Assignees   []User       `json:"assignees"`
}

type Status struct {
	Status string `json:"status"`
}

type Priority struct {
	Priority string `json:"priority"`
}

type Checklist struct {
	Name  string          `json:"name" yaml:"name"`
	Items []ChecklistItem `json:"items" yaml:"items"`
}

type ChecklistItem struct {
	Name     string `json:"name" yaml:"name"`
	Resolved bool   `json:"resolved" yaml:"resolved"`
}

type Attachment struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type Comment struct {
	ID          string `json:"id"`
	CommentText string `json:"comment_text"`
	User        User   `json:"user"`
	Date        string `json:"date"`
}

func (p *Priority) String() string {
	if p == nil {
		return ""
	}
	return p.Priority
}
