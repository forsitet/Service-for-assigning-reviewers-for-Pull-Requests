package service

import (
	"context"
	"database/sql"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/domain"
	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/repo/postgres"
	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/service"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	connString := os.Getenv("TEST_DATABASE_CONN")
	if connString == "" {
		t.Skip("TEST_DATABASE_CONN is not set, skipping integration tests")
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	db, err := postgres.NewDB(ctx, connString, logger)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	if err := postgres.RunMigrations(ctx, db, logger); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

func cleanupTables(t *testing.T, db *sql.DB) {
	t.Helper()

	ctx := context.Background()
	queries := []string{
		`DELETE FROM pull_request_reviewers`,
		`DELETE FROM pull_requests`,
		`DELETE FROM users`,
		`DELETE FROM teams`,
	}

	for _, q := range queries {
		if _, err := db.ExecContext(ctx, q); err != nil {
			t.Fatalf("cleanup query %q failed: %v", q, err)
		}
	}
}

func TestPCreateAndMerge(t *testing.T) {
	db := openTestDB(t)
	cleanupTables(t, db)

	ctx := context.Background()

	seedSQL := `
INSERT INTO teams(name) VALUES ('backend');

INSERT INTO users(id, username, team_name, is_active)
VALUES
  ('u1', 'Alice', 'backend', true),
  ('u2', 'Bob',   'backend', true),
  ('u3', 'Carol', 'backend', true);
`
	if _, err := db.ExecContext(ctx, seedSQL); err != nil {
		t.Fatalf("seed data failed: %v", err)
	}

	userRepo := postgres.NewUserRepo(db)
	prRepo := postgres.NewPRRepo(db)

	fixedTime := time.Unix(1_700_000_000, 0)
	nowFunc := func() time.Time { return fixedTime }

	svc := service.NewPRService(prRepo, userRepo, nowFunc)

	pr, err := svc.CreatePullRequest(ctx, "pr-1", "Test PR", "u1")
	if err != nil {
		t.Fatalf("CreatePullRequest returned error: %v", err)
	}

	if pr.ID != "pr-1" {
		t.Fatalf("expected ID pr-1, got %s", pr.ID)
	}
	if pr.Name != "Test PR" {
		t.Fatalf("expected Name Test PR, got %s", pr.Name)
	}
	if pr.AuthorID != "u1" {
		t.Fatalf("expected AuthorID u1, got %s", pr.AuthorID)
	}
	if pr.Status != domain.PRStatusOpen {
		t.Fatalf("expected Status OPEN, got %s", pr.Status)
	}
	if pr.CreatedAt != fixedTime.Unix() {
		t.Fatalf("expected CreatedAt %d, got %d", fixedTime.Unix(), pr.CreatedAt)
	}
	if len(pr.AssignedReviewers) != 2 {
		t.Fatalf("expected 2 reviewers, got %d", len(pr.AssignedReviewers))
	}
	for _, r := range pr.AssignedReviewers {
		if r == "u1" {
			t.Fatalf("author u1 must not be in AssignedReviewers")
		}
	}

	var (
		dbStatus   string
		dbCreated  time.Time
		dbMergedAt sql.NullTime
	)

	row := db.QueryRowContext(ctx,
		`SELECT status, created_at, merged_at FROM pull_requests WHERE id = $1`,
		"pr-1",
	)
	if err := row.Scan(&dbStatus, &dbCreated, &dbMergedAt); err != nil {
		t.Fatalf("select created PR from DB failed: %v", err)
	}

	if dbStatus != "OPEN" {
		t.Fatalf("expected DB status OPEN, got %s", dbStatus)
	}
	if dbMergedAt.Valid {
		t.Fatalf("expected merged_at to be NULL for OPEN PR")
	}

	mergedPR, err := svc.MergePullRequest(ctx, "pr-1")
	if err != nil {
		t.Fatalf("MergePullRequest returned error: %v", err)
	}

	if mergedPR.Status != domain.PRStatusMerged {
		t.Fatalf("expected Status MERGED, got %s", mergedPR.Status)
	}
	if mergedPR.MergedAt == 0 {
		t.Fatalf("expected mergedPR.MergedAt not nil")
	}

	row = db.QueryRowContext(ctx,
		`SELECT status, merged_at FROM pull_requests WHERE id = $1`,
		"pr-1",
	)

	dbMergedAt = sql.NullTime{}
	if err := row.Scan(&dbStatus, &dbMergedAt); err != nil {
		t.Fatalf("select merged PR from DB failed: %v", err)
	}

	if dbStatus != "MERGED" {
		t.Fatalf("expected DB status MERGED, got %s", dbStatus)
	}
	if !dbMergedAt.Valid {
		t.Fatalf("expected merged_at to be NOT NULL after merge")
	}
}
