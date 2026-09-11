package clickup

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const baseURL = "https://api.clickup.com/api/v2"

type Client struct {
	apiToken string
	http     *http.Client
}

func NewClient(apiToken string) *Client {
	return &Client{
		apiToken: apiToken,
		http:     &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) doGet(path string, query url.Values, out any) error {
	u := baseURL + path
	if query != nil {
		u += "?" + query.Encode()
	}
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.apiToken)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request %s: %w", path, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response %s: %w", path, err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("clickup api %s returned %d: %s", path, resp.StatusCode, string(body))
	}
	if out != nil {
		if err := json.Unmarshal(body, out); err != nil {
			return fmt.Errorf("decode response %s: %w", path, err)
		}
	}
	return nil
}

// GetTeamTasksByAssignee fetches every task across the given ClickUp workspace
// (team) assigned to the given ClickUp user ID.
func (c *Client) GetTeamTasksByAssignee(teamID, clickupUserID string) ([]Task, error) {
	query := url.Values{}
	query.Add("assignees[]", clickupUserID)
	query.Add("include_closed", "true")

	var out struct {
		Tasks []Task `json:"tasks"`
	}
	if err := c.doGet(fmt.Sprintf("/team/%s/task", teamID), query, &out); err != nil {
		return nil, err
	}
	return out.Tasks, nil
}

func (c *Client) GetTaskComments(taskID string) ([]Comment, error) {
	var out struct {
		Comments []Comment `json:"comments"`
	}
	if err := c.doGet(fmt.Sprintf("/task/%s/comment", taskID), nil, &out); err != nil {
		return nil, err
	}
	return out.Comments, nil
}

// DownloadAttachment fetches the raw bytes of a ClickUp attachment URL, using
// the same API token for auth.
func (c *Client) DownloadAttachment(attachmentURL string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, attachmentURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.apiToken)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download attachment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("download attachment returned %d: %s", resp.StatusCode, string(body))
	}
	return io.ReadAll(resp.Body)
}
