# GoTask

A version-controlled task ledger that syncs ClickUp tasks into GitHub, and sends
WhatsApp status reminders via `wacli` (ProjectPulse).

## Architecture

- **ClickUp connector** — fetches current task state from the ClickUp API on a
  scheduled reconciliation run (twice daily).
- **Local DB (SQLite)** — holds the assignee registry (ClickUp user → notification
  channel details → GitHub branch name) and a snapshot cache of each task's last-known
  field values, used only to compute diffs each run. Git is the historical record;
  the DB is not.
- **GitHub writer** — commits diffs to the right branch. One commit per detected
  change. No diff, no commit, no push.
- **Notification dispatcher** — generic `notify(assignee, event_type, payload)`
  interface with a pluggable channel registry. Only a `whatsapp` module (wrapping
  `wacli`) is implemented for now; Discord/Telegram can be added later behind the
  same interface.
- **Orchestrator** — runs the pipeline twice daily: fetch → diff against DB →
  update DB → write GitHub commits → dispatch notifications.

## Branching model

One branch per assignee. A branch contains **only** that person's task directories —
branches are not meant to be merged into each other or into `main`. `main` holds this
repo's own code/docs/config, not task data.

### Task directory layout (per branch)

```
task-<clickup_id>/
  task.yaml            # scalar fields: id, title, status, priority, due_date,
                        # assignee, created_at, updated_at, provenance, attachments: [...]
  description.md       # task description body
  checklist.yaml        # subtasks/checklist items, if any
  comments.md           # all comments, appended over time
  attachments/
    <original-filename>
```

### Reassignment

When a task's assignee changes, three commits record the move:

1. **Relabel** (on the source branch) — `task.yaml` updated to note the reassignment
   and destination.
2. **Delete** (on the source branch) — the task directory is removed.
3. **Add** (on the destination branch) — the task directory is recreated with a
   `provenance` block in `task.yaml`:

   ```yaml
   provenance:
     reassigned_from: alice
     source_commit: <sha of the relabel commit on alice's branch>
     reassigned_at: 2026-09-12T10:00:00Z
   ```

Both the old and new assignee receive a WhatsApp notification on reassignment.

### Change detection

Every reconciliation run diffs ClickUp's current state against the DB snapshot for
each task. Any difference (status, due date, priority, description, comments,
attachments, checklist — anything) produces exactly one commit and one WhatsApp
change-alert to the assignee. The morning run additionally sends each assignee a
daily digest of their currently open tasks.

The sync is idempotent per task: each run compares the task's actual current branch
against where ClickUp says it should be, and only performs whichever of
relabel/delete/add hasn't happened yet — so a crash mid-move gets completed (or
safely retried) on the next run rather than leaving an orphaned task.
