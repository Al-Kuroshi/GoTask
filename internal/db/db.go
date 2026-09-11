package db

import (
	"database/sql"
	_ "embed"
	"fmt"

	_ "modernc.org/sqlite"
)

//go:embed migrations.sql
var migrations string

func Open(path string) (*sql.DB, error) {
	database, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if _, err := database.Exec(migrations); err != nil {
		database.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}
	return database, nil
}
