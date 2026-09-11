CREATE TABLE IF NOT EXISTS assignees (
  id                INTEGER PRIMARY KEY AUTOINCREMENT,
  clickup_user_id   TEXT NOT NULL UNIQUE,
  clickup_username  TEXT NOT NULL,
  github_username   TEXT,
  whatsapp_number   TEXT,
  branch_name       TEXT NOT NULL UNIQUE,
  sync_status       TEXT NOT NULL DEFAULT 'pending',
  sync_error        TEXT,
  last_synced_at    DATETIME,
  created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
