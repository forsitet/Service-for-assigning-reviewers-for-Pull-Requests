package postgres

import (
	"context"
	"fmt"
)

func (r *PRRepo) CountAssignmentsByReviewer(ctx context.Context) (map[string]int64, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT reviewer_id, COUNT(*) AS cnt
         FROM pull_request_reviewers
         GROUP BY reviewer_id
         ORDER BY reviewer_id`,
	)
	if err != nil {
		return nil, fmt.Errorf("count assignments by reviewer: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// Log error but don't fail - rows are already read
		}
	}()

	result := make(map[string]int64)
	for rows.Next() {
		var (
			id  string
			cnt int64
		)
		if err := rows.Scan(&id, &cnt); err != nil {
			return nil, fmt.Errorf("scan assignments by reviewer: %w", err)
		}
		result[id] = cnt
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate assignments by reviewer: %w", err)
	}

	return result, nil
}

func (r *PRRepo) CountAssignmentsByPR(ctx context.Context) (map[string]int64, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT pr_id, COUNT(*) AS cnt
         FROM pull_request_reviewers
         GROUP BY pr_id
         ORDER BY pr_id`,
	)
	if err != nil {
		return nil, fmt.Errorf("count assignments by pr: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// Log error but don't fail - rows are already read
		}
	}()

	result := make(map[string]int64)
	for rows.Next() {
		var (
			id  string
			cnt int64
		)
		if err := rows.Scan(&id, &cnt); err != nil {
			return nil, fmt.Errorf("scan assignments by pr: %w", err)
		}
		result[id] = cnt
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate assignments by pr: %w", err)
	}

	return result, nil
}
