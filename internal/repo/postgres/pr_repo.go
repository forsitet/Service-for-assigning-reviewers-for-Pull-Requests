package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/domain"
)

type PRRepo struct {
	db *sql.DB
}

func NewPRRepo(db *sql.DB) *PRRepo {
	return &PRRepo{db: db}
}

func (r *PRRepo) CreateWithReviewers(ctx context.Context, pr *domain.PullRequest, reviewerIDs []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin create PR tx: %w", err)
	}
	defer func() {
		// #nosec G104 -- error is ignored in defer rollback
		_ = tx.Rollback()
	}()

	var createdAt *time.Time
	if pr.CreatedAt != 0 {
		t := time.Unix(pr.CreatedAt, 0)
		createdAt = &t
	}
	var mergedAt *time.Time
	if pr.MergedAt != 0 {
		t := time.Unix(pr.MergedAt, 0)
		mergedAt = &t
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO pull_requests (id, name, author_id, status, created_at, merged_at)
         VALUES ($1, $2, $3, $4, COALESCE($5, now()), $6)`,
		pr.ID,
		pr.Name,
		pr.AuthorID,
		string(pr.Status),
		createdAt,
		mergedAt,
	)
	if err != nil {
		return fmt.Errorf("insert pull_request: %w", err)
	}

	for _, reviewerID := range reviewerIDs {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO pull_request_reviewers (pr_id, reviewer_id)
             VALUES ($1, $2)`,
			pr.ID, reviewerID,
		); err != nil {
			return fmt.Errorf("insert pull_request_reviewer %s: %w", reviewerID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit create PR tx: %w", err)
	}

	return nil
}

func (r *PRRepo) GetByID(ctx context.Context, id string) (*domain.PullRequest, []string, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, author_id, status, created_at, merged_at
         FROM pull_requests
         WHERE id = $1`,
		id,
	)

	var (
		prID       string
		name       string
		authorID   string
		statusStr  string
		createdRaw sql.NullTime
		mergedRaw  sql.NullTime
	)

	if err := row.Scan(&prID, &name, &authorID, &statusStr, &createdRaw, &mergedRaw); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, sql.ErrNoRows
		}
		return nil, nil, fmt.Errorf("get pull_request: %w", err)
	}

	var createdAt int64
	if createdRaw.Valid {
		createdAt = createdRaw.Time.Unix()
	}
	var mergedAt int64
	if mergedRaw.Valid {
		mergedAt = mergedRaw.Time.Unix()
	}

	pr := &domain.PullRequest{
		ID:                prID,
		Name:              name,
		AuthorID:          authorID,
		Status:            domain.PRStatus(statusStr),
		AssignedReviewers: nil,
		CreatedAt:         createdAt,
		MergedAt:          mergedAt,
	}

	reviewers, err := r.loadReviewers(ctx, prID)
	if err != nil {
		return nil, nil, err
	}
	pr.AssignedReviewers = reviewers

	return pr, reviewers, nil
}

func (r *PRRepo) SetMerged(ctx context.Context, id string, mergedAt time.Time) (*domain.PullRequest, []string, error) {
	row := r.db.QueryRowContext(ctx,
		`UPDATE pull_requests
         SET status = 'MERGED',
             merged_at = $2
         WHERE id = $1
         RETURNING id, name, author_id, status, created_at, merged_at`,
		id, mergedAt,
	)

	var (
		prID       string
		name       string
		authorID   string
		statusStr  string
		createdRaw sql.NullTime
		mergedRaw  sql.NullTime
	)

	if err := row.Scan(&prID, &name, &authorID, &statusStr, &createdRaw, &mergedRaw); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, sql.ErrNoRows
		}
		return nil, nil, fmt.Errorf("set merged: %w", err)
	}

	var createdAt int64
	if createdRaw.Valid {
		createdAt = createdRaw.Time.Unix()
	}
	var mergedAtUnix int64
	if mergedRaw.Valid {
		mergedAtUnix = mergedRaw.Time.Unix()
	}

	pr := &domain.PullRequest{
		ID:                prID,
		Name:              name,
		AuthorID:          authorID,
		Status:            domain.PRStatus(statusStr),
		AssignedReviewers: nil,
		CreatedAt:         createdAt,
		MergedAt:          mergedAtUnix,
	}

	reviewers, err := r.loadReviewers(ctx, prID)
	if err != nil {
		return nil, nil, err
	}
	pr.AssignedReviewers = reviewers

	return pr, reviewers, nil
}

func (r *PRRepo) UpdateReviewers(ctx context.Context, id string, reviewerIDs []string) (*domain.PullRequest, []string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("begin update reviewers tx: %w", err)
	}
	defer func() {
		// #nosec G104 -- error is ignored in defer rollback
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM pull_request_reviewers WHERE pr_id = $1`,
		id,
	); err != nil {
		return nil, nil, fmt.Errorf("delete old reviewers: %w", err)
	}

	for _, reviewerID := range reviewerIDs {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO pull_request_reviewers (pr_id, reviewer_id)
             VALUES ($1, $2)`,
			id, reviewerID,
		); err != nil {
			return nil, nil, fmt.Errorf("insert new reviewer %s: %w", reviewerID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, fmt.Errorf("commit update reviewers tx: %w", err)
	}

	pr, reviewers, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return pr, reviewers, nil
}

func (r *PRRepo) ListByReviewer(ctx context.Context, userID string) ([]domain.PullRequest, error) {
	dbRows, err := r.db.QueryContext(ctx,
		`SELECT p.id, p.name, p.author_id, p.status, p.created_at, p.merged_at
         FROM pull_requests p
         INNER JOIN pull_request_reviewers r
             ON p.id = r.pr_id
         WHERE r.reviewer_id = $1
         ORDER BY p.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list pull_requests by reviewer: %w", err)
	}
	defer func() {
		if err := dbRows.Close(); err != nil {
			// Log error but don't fail - rows are already read
		}
	}()

	result := make([]domain.PullRequest, 0)
	for dbRows.Next() {
		var (
			prID       string
			name       string
			authorID   string
			statusStr  string
			createdRaw sql.NullTime
			mergedRaw  sql.NullTime
		)

		if err := dbRows.Scan(&prID, &name, &authorID, &statusStr, &createdRaw, &mergedRaw); err != nil {
			return nil, fmt.Errorf("scan pull_request: %w", err)
		}

		var createdAt int64
		if createdRaw.Valid {
			createdAt = createdRaw.Time.Unix()
		}
		var mergedAt int64
		if mergedRaw.Valid {
			mergedAt = mergedRaw.Time.Unix()
		}

		pr := domain.PullRequest{
			ID:        prID,
			Name:      name,
			AuthorID:  authorID,
			Status:    domain.PRStatus(statusStr),
			CreatedAt: createdAt,
			MergedAt:  mergedAt,
		}
		result = append(result, pr)
	}
	if err := dbRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pull_requests by reviewer: %w", err)
	}

	return result, nil
}

func (r *PRRepo) loadReviewers(ctx context.Context, prID string) ([]string, error) {
	dbRows, err := r.db.QueryContext(ctx,
		`SELECT reviewer_id
         FROM pull_request_reviewers
         WHERE pr_id = $1
         ORDER BY reviewer_id`,
		prID,
	)
	if err != nil {
		return nil, fmt.Errorf("list reviewers: %w", err)
	}
	defer func() {
		if err := dbRows.Close(); err != nil {
			// Log error but don't fail - rows are already read
		}
	}()

	reviewers := make([]string, 0)
	for dbRows.Next() {
		var id string
		if err := dbRows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan reviewer_id: %w", err)
		}
		reviewers = append(reviewers, id)
	}
	if err := dbRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reviewers: %w", err)
	}

	return reviewers, nil
}

func (r *PRRepo) Exists(ctx context.Context, id string) (bool, error) {
	var dummy string
	err := r.db.QueryRowContext(ctx,
		`SELECT id FROM pull_requests WHERE id = $1`,
		id,
	).Scan(&dummy)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("check PR exists: %w", err)
	}
	return true, nil
}
