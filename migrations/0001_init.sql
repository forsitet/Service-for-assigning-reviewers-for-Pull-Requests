-- +goose Up
CREATE TABLE IF NOT EXISTS teams (
  name TEXT PRIMARY KEY,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY,
  username TEXT NOT NULL,
  team_name TEXT NOT NULL REFERENCES teams(name) ON DELETE RESTRICT,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS pull_requests (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  author_id TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  merged_at TIMESTAMPTZ
);
CREATE TABLE IF NOT EXISTS pull_request_reviewers (
  pr_id TEXT NOT NULL REFERENCES pull_requests(id) ON DELETE CASCADE,
  reviewer_id TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  PRIMARY KEY (pr_id, reviewer_id)
);
-- +goose Down
DROP TABLE IF EXISTS pull_request_reviewers;
DROP TABLE IF EXISTS pull_requests;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;
