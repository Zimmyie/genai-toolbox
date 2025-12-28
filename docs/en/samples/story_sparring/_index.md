---
title: "Story writing sparring partner"
type: docs
weight: 10
description: >
  Use Toolbox + SQLite to give your model durable story context and a focused,
  high-energy writing partner.
---

## Overview

If your writing sessions are long, relying on the model's short-term context
window can lead to drift. This sample uses a persistent SQLite database as
"extra storage" so your assistant can recall characters, scene goals, and
open threads between sessions. You'll also give the model a clear system prompt
so it stays on topic, matches your energy, and offers specific suggestions
without echoing your words.

## 1) Create a local story database

Create a SQLite file (this lives on disk, so it survives restarts):

```bash
sqlite3 ./story.db <<'SQL'
CREATE TABLE IF NOT EXISTS projects (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  title TEXT NOT NULL,
  logline TEXT,
  tone TEXT,
  created_at TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS characters (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  project_id INTEGER NOT NULL,
  name TEXT NOT NULL,
  role TEXT,
  summary TEXT,
  arc TEXT,
  FOREIGN KEY (project_id) REFERENCES projects(id)
);

CREATE TABLE IF NOT EXISTS scenes (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  project_id INTEGER NOT NULL,
  scene_number INTEGER,
  title TEXT,
  goal TEXT,
  conflict TEXT,
  outcome TEXT,
  notes TEXT,
  updated_at TEXT DEFAULT (datetime('now')),
  FOREIGN KEY (project_id) REFERENCES projects(id)
);

CREATE TABLE IF NOT EXISTS threads (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  project_id INTEGER NOT NULL,
  description TEXT NOT NULL,
  status TEXT DEFAULT 'open',
  updated_at TEXT DEFAULT (datetime('now')),
  FOREIGN KEY (project_id) REFERENCES projects(id)
);
SQL
```

## 2) Configure Toolbox with SQLite tools

Create a `tools.yaml` file that points at your database and exposes focused
queries for story work:

```yaml
sources:
  story-db:
    kind: sqlite
    database: ./story.db

tools:
  story-create-project:
    kind: sqlite-sql
    source: story-db
    description: Create a new story project.
    parameters:
      - name: title
        type: string
      - name: logline
        type: string
      - name: tone
        type: string
    statement: |
      INSERT INTO projects (title, logline, tone)
      VALUES (?, ?, ?);

  story-add-character:
    kind: sqlite-sql
    source: story-db
    description: Add a character to a project.
    parameters:
      - name: project_id
        type: integer
      - name: name
        type: string
      - name: role
        type: string
      - name: summary
        type: string
      - name: arc
        type: string
    statement: |
      INSERT INTO characters (project_id, name, role, summary, arc)
      VALUES (?, ?, ?, ?, ?);

  story-upsert-scene:
    kind: sqlite-sql
    source: story-db
    description: Add or update a scene's goal/conflict/outcome notes.
    parameters:
      - name: project_id
        type: integer
      - name: scene_number
        type: integer
      - name: title
        type: string
      - name: goal
        type: string
      - name: conflict
        type: string
      - name: outcome
        type: string
      - name: notes
        type: string
    statement: |
      INSERT INTO scenes (project_id, scene_number, title, goal, conflict, outcome, notes)
      VALUES (?, ?, ?, ?, ?, ?, ?)
      ON CONFLICT(project_id, scene_number)
      DO UPDATE SET
        title=excluded.title,
        goal=excluded.goal,
        conflict=excluded.conflict,
        outcome=excluded.outcome,
        notes=excluded.notes,
        updated_at=datetime('now');

  story-open-threads:
    kind: sqlite-sql
    source: story-db
    description: List open story threads that still need resolution.
    parameters:
      - name: project_id
        type: integer
    statement: |
      SELECT id, description, status, updated_at
      FROM threads
      WHERE project_id = ? AND status = 'open'
      ORDER BY updated_at DESC;

  story-summary:
    kind: sqlite-sql
    source: story-db
    description: Pull a project summary with characters and recent scenes.
    parameters:
      - name: project_id
        type: integer
    statement: |
      SELECT p.id, p.title, p.logline, p.tone FROM projects p WHERE p.id = ?;

      SELECT name, role, summary, arc
      FROM characters
      WHERE project_id = ?
      ORDER BY name;

      SELECT scene_number, title, goal, conflict, outcome, notes, updated_at
      FROM scenes
      WHERE project_id = ?
      ORDER BY scene_number;
```

> **Note:** If you want to use the `ON CONFLICT` upsert above, add a unique
> constraint:
>
> ```sql
> CREATE UNIQUE INDEX IF NOT EXISTS scenes_project_scene
> ON scenes(project_id, scene_number);
> ```

## 3) Run Toolbox

```bash
toolbox --tools-file ./tools.yaml
```

Now your model can call tools like `story-summary` to reload the state of your
story at any time, which keeps it on topic even across long sessions.

## 4) Use a focused system prompt

Pair the tools with a system message that defines the assistant's role as a
sparring partner. This keeps the feedback energetic, specific, and not a
mirror of your wording:

```
You are my story-writing sparring partner. Stay on topic and match my energy.
Do not echo my text back verbatim. Give concrete suggestions, improvements, or
next steps. If a detail is missing, ask a short, pointed question. Use the
story tools to recall characters, scenes, and open threads before advising.
```

## 5) Suggested workflow

1. Start each session by calling `story-summary` for your active project.
2. Add or update scenes as you write using `story-upsert-scene`.
3. Track unresolved beats with `threads` and check them using `story-open-threads`.

With this flow, your model has durable storage for your screenplay and can
act like a focused collaborator instead of a generic suggestion engine.
