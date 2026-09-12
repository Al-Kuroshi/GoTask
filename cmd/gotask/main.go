package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Al-Kuroshi/GoTask/internal/clickup"
	"github.com/Al-Kuroshi/GoTask/internal/config"
	"github.com/Al-Kuroshi/GoTask/internal/db"
	"github.com/Al-Kuroshi/GoTask/internal/gitrepo"
	"github.com/Al-Kuroshi/GoTask/internal/sync"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "register-assignee":
		cmdRegisterAssignee(os.Args[2:])
	case "sync-assignee":
		cmdSyncAssignee(os.Args[2:])
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: gotask register-assignee -clickup-user-id ID -clickup-username NAME [-github-user NAME] [-whatsapp NUMBER] [-config PATH]")
	fmt.Fprintln(os.Stderr, "       gotask sync-assignee -assignee-id ID [-config PATH]")
}

func cmdSyncAssignee(args []string) {
	fs := flag.NewFlagSet("sync-assignee", flag.ExitOnError)
	assigneeID := fs.Int64("assignee-id", 0, "assignee DB id (required)")
	configPath := fs.String("config", "config.yaml", "path to config file")
	fs.Parse(args)

	if *assigneeID == 0 {
		fmt.Fprintln(os.Stderr, "error: -assignee-id is required")
		os.Exit(1)
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		fatal(err)
	}
	database, err := db.Open(cfg.DB.Path)
	if err != nil {
		fatal(err)
	}
	defer database.Close()

	token, err := cfg.GitHubToken()
	if err != nil {
		fatal(err)
	}
	repo := gitrepo.New(cfg.GitHub.RepoPath, token)
	cu := clickup.NewClient(cfg.ClickUp.APIToken)

	if err := sync.SyncAssignee(database, repo, cu, cfg.ClickUp.TeamID, *assigneeID); err != nil {
		fmt.Fprintf(os.Stderr, "sync finished with error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Sync complete.")
}

func cmdRegisterAssignee(args []string) {
	fs := flag.NewFlagSet("register-assignee", flag.ExitOnError)
	clickupUserID := fs.String("clickup-user-id", "", "ClickUp numeric user ID (required)")
	clickupUsername := fs.String("clickup-username", "", "ClickUp username (required)")
	githubUsername := fs.String("github-user", "", "GitHub username (optional)")
	whatsapp := fs.String("whatsapp", "", "WhatsApp number, e.g. +15551234567 (optional)")
	configPath := fs.String("config", "config.yaml", "path to config file")
	fs.Parse(args)

	if *clickupUserID == "" || *clickupUsername == "" {
		fmt.Fprintln(os.Stderr, "error: -clickup-user-id and -clickup-username are required")
		os.Exit(1)
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		fatal(err)
	}

	database, err := db.Open(cfg.DB.Path)
	if err != nil {
		fatal(err)
	}
	defer database.Close()

	token, err := cfg.GitHubToken()
	if err != nil {
		fatal(err)
	}
	repo := gitrepo.New(cfg.GitHub.RepoPath, token)

	id, err := sync.RegisterAssignee(database, repo, sync.RegisterParams{
		ClickUpUserID:   *clickupUserID,
		ClickUpUsername: *clickupUsername,
		GitHubUsername:  *githubUsername,
		WhatsAppNumber:  *whatsapp,
	})
	if err != nil {
		fatal(err)
	}
	fmt.Printf("Registered assignee id=%d, branch created.\n", id)

	cu := clickup.NewClient(cfg.ClickUp.APIToken)
	if err := sync.SyncAssignee(database, repo, cu, cfg.ClickUp.TeamID, id); err != nil {
		fmt.Fprintf(os.Stderr, "sync finished with error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Sync complete.")
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
