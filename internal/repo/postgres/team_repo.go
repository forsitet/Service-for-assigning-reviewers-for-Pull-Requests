package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/domain"
)

type TeamRepo struct {
	db *sql.DB
}

func NewTeamRepo(db *sql.DB) *TeamRepo {
	return &TeamRepo{db: db}
}

func (r *TeamRepo) Create(ctx context.Context, name string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO teams (name) VALUES ($1)`,
		name,
	)
	if err != nil {
		return fmt.Errorf("insert team: %w", err)
	}
	return nil
}

func (r *TeamRepo) Exists(ctx context.Context, name string) (bool, error) {
	var dummy int
	err := r.db.QueryRowContext(ctx,
		`SELECT 1 FROM teams WHERE name = $1`,
		name,
	).Scan(&dummy)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("check team exists: %w", err)
	}
	return true, nil
}

func (r *TeamRepo) GetWithMembers(ctx context.Context, name string) (*domain.Team, error) {
	var teamName string
	err := r.db.QueryRowContext(ctx,
		`SELECT name FROM teams WHERE name = $1`,
		name,
	).Scan(&teamName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("get team: %w", err)
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, username, team_name, is_active
         FROM users
         WHERE team_name = $1
         ORDER BY id`,
		name,
	)
	if err != nil {
		return nil, fmt.Errorf("list team members: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// Log error but don't fail - rows are already read
		}
	}()

	members := make([]domain.User, 0)
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		members = append(members, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate team members: %w", err)
	}

	return &domain.Team{
		Name:    teamName,
		Members: members,
	}, nil
}
