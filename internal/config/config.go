package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ClickUp struct {
		APIToken string `yaml:"api_token"`
		TeamID   string `yaml:"team_id"`
	} `yaml:"clickup"`
	GitHub struct {
		RepoPath string `yaml:"repo_path"`
		TokenEnv string `yaml:"token_env"`
	} `yaml:"github"`
	DB struct {
		Path string `yaml:"path"`
	} `yaml:"db"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	if err := c.validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *Config) validate() error {
	switch {
	case c.ClickUp.APIToken == "":
		return fmt.Errorf("clickup.api_token is required")
	case c.ClickUp.TeamID == "":
		return fmt.Errorf("clickup.team_id is required")
	case c.GitHub.RepoPath == "":
		return fmt.Errorf("github.repo_path is required")
	case c.GitHub.TokenEnv == "":
		return fmt.Errorf("github.token_env is required")
	case c.DB.Path == "":
		return fmt.Errorf("db.path is required")
	}
	return nil
}

// GitHubToken reads the GitHub PAT from the environment variable named by
// github.token_env, rather than storing it directly in the config file.
func (c *Config) GitHubToken() (string, error) {
	token := os.Getenv(c.GitHub.TokenEnv)
	if token == "" {
		return "", fmt.Errorf("environment variable %s is not set", c.GitHub.TokenEnv)
	}
	return token, nil
}
