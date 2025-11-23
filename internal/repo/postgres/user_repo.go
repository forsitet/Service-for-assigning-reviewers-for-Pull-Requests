package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/domain"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) UpsertForTeam(ctx context.Context, teamName string, users []domain.User) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin upsert users tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	stmt := `
INSERT INTO users (id, username, team_name, is_active)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE
SET username = EXCLUDED.username,
    team_name = EXCLUDED.team_name,
    is_active = EXCLUDED.is_active,
    updated_at = now()
`
	for _, u := range users {
		if _, err := tx.ExecContext(ctx, stmt,
			u.ID,
			u.Username,
			teamName,
			u.IsActive,
		); err != nil {
			return fmt.Errorf("upsert user %s: %w", u.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit upsert users tx: %w", err)
	}

	return nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var u domain.User
	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, team_name, is_active
         FROM users
         WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &u, nil
}

func (r *UserRepo) SetIsActive(ctx context.Context, id string, active bool) (*domain.User, error) {
	var u domain.User
	err := r.db.QueryRowContext(ctx,
		`UPDATE users
         SET is_active = $2,
             updated_at = now()
         WHERE id = $1
         RETURNING id, username, team_name, is_active`,
		id, active,
	).Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("set user is_active: %w", err)
	}
	return &u, nil
}

func (r *UserRepo) ListByTeam(ctx context.Context, teamName string) ([]domain.User, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, username, team_name, is_active
         FROM users
         WHERE team_name = $1
         ORDER BY id`,
		teamName,
	)
	if err != nil {
		return nil, fmt.Errorf("list users by team: %w", err)
	}
	defer rows.Close()

	users := make([]domain.User, 0)
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users by team: %w", err)
	}

	return users, nil
}
