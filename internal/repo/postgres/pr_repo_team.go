package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/forsitet/Service-for-assigning-reviewers-for-Pull-Requests/internal/domain"
)

func (r *PRRepo) DeactivateTeamAndReassignOpenPRs(ctx context.Context, teamName string) (domain.TeamDeactivationResult, error) {
	result := domain.TeamDeactivationResult{
		TeamName: teamName,
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return result, fmt.Errorf("begin tx deactivate team %s: %w", teamName, err)
	}
	defer func() {
		// #nosec G104 -- error is ignored in defer rollback
		_ = tx.Rollback()
	}()

	if err := r.checkTeamExists(ctx, tx, teamName); err != nil {
		return result, err
	}

	deactivatedIDs, err := r.deactivateUsers(ctx, tx, teamName)
	if err != nil {
		return result, err
	}

	result.DeactivatedUsers = len(deactivatedIDs)
	if len(deactivatedIDs) == 0 {
		if err := tx.Commit(); err != nil {
			return result, fmt.Errorf("commit empty deactivation tx: %w", err)
		}
		return result, nil
	}

	prMap, err := r.loadAffectedPRs(ctx, tx, deactivatedIDs)
	if err != nil {
		return result, err
	}

	if len(prMap) == 0 {
		if err := tx.Commit(); err != nil {
			return result, fmt.Errorf("commit tx without pr updates: %w", err)
		}
		return result, nil
	}

	if err := r.loadCurrentReviewers(ctx, tx, prMap); err != nil {
		return result, err
	}

	authorTeam, err := r.loadAuthorTeams(ctx, tx, prMap)
	if err != nil {
		return result, err
	}

	candidatesByTeam, err := r.loadCandidates(ctx, tx, authorTeam)
	if err != nil {
		return result, err
	}

	newReviewersByPR := r.calculateNewReviewers(prMap, authorTeam, candidatesByTeam)

	if err := r.updatePRReviewers(ctx, tx, newReviewersByPR); err != nil {
		return result, err
	}

	result.UpdatedPullRequests = len(newReviewersByPR)

	if err := tx.Commit(); err != nil {
		return result, fmt.Errorf("commit deactivate team tx: %w", err)
	}

	return result, nil
}

type prInfo struct {
	authorID    string
	deactivated map[string]struct{}
	current     []string
}

func (r *PRRepo) checkTeamExists(ctx context.Context, tx *sql.Tx, teamName string) error {
	var tmp string
	if err := tx.QueryRowContext(ctx,
		`SELECT name FROM teams WHERE name = $1`,
		teamName,
	).Scan(&tmp); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.NewDomainError(domain.ErrorCodeNotFound, "team not found")
		}
		return fmt.Errorf("check team exists: %w", err)
	}
	return nil
}

func (r *PRRepo) deactivateUsers(ctx context.Context, tx *sql.Tx, teamName string) ([]string, error) {
	deactivatedIDs := make([]string, 0)
	rows, err := tx.QueryContext(ctx,
		`UPDATE users
         SET is_active = false,
             updated_at = now()
         WHERE team_name = $1 AND is_active = true
         RETURNING id`,
		teamName,
	)
	if err != nil {
		return nil, fmt.Errorf("deactivate users of team %s: %w", teamName, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// Log error but don't fail - rows are already read
		}
	}()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan deactivated user id: %w", err)
		}
		deactivatedIDs = append(deactivatedIDs, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate deactivated user ids: %w", err)
	}
	return deactivatedIDs, nil
}

func (r *PRRepo) loadAffectedPRs(ctx context.Context, tx *sql.Tx, deactivatedIDs []string) (map[string]*prInfo, error) {
	query, args := buildInClause(`
        SELECT p.id, p.author_id, r.reviewer_id
        FROM pull_request_reviewers r
        JOIN pull_requests p ON p.id = r.pr_id
        WHERE p.status = 'OPEN' AND r.reviewer_id IN (`, deactivatedIDs)

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("select affected pull requests: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// Log error but don't fail - rows are already read
		}
	}()

	prMap := make(map[string]*prInfo)
	for rows.Next() {
		var (
			prID       string
			authorID   string
			reviewerID string
		)
		if err := rows.Scan(&prID, &authorID, &reviewerID); err != nil {
			return nil, fmt.Errorf("scan affected pr row: %w", err)
		}

		info, ok := prMap[prID]
		if !ok {
			info = &prInfo{
				authorID:    authorID,
				deactivated: make(map[string]struct{}),
				current:     make([]string, 0, 2),
			}
			prMap[prID] = info
		}
		info.deactivated[reviewerID] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate affected prs: %w", err)
	}
	return prMap, nil
}

func (r *PRRepo) loadCurrentReviewers(ctx context.Context, tx *sql.Tx, prMap map[string]*prInfo) error {
	prIDs := make([]string, 0, len(prMap))
	for prID := range prMap {
		prIDs = append(prIDs, prID)
	}

	query, args := buildInClause(`
        SELECT pr_id, reviewer_id
        FROM pull_request_reviewers
        WHERE pr_id IN (`, prIDs)

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("load current reviewers: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// Log error but don't fail - rows are already read
		}
	}()

	for rows.Next() {
		var (
			prID       string
			reviewerID string
		)
		if err := rows.Scan(&prID, &reviewerID); err != nil {
			return fmt.Errorf("scan current reviewer: %w", err)
		}
		info, ok := prMap[prID]
		if !ok {
			continue
		}
		info.current = append(info.current, reviewerID)
	}
	return rows.Err()
}

func (r *PRRepo) loadAuthorTeams(ctx context.Context, tx *sql.Tx, prMap map[string]*prInfo) (map[string]string, error) {
	authorSet := make(map[string]struct{})
	for _, info := range prMap {
		authorSet[info.authorID] = struct{}{}
	}

	authorIDs := make([]string, 0, len(authorSet))
	for id := range authorSet {
		authorIDs = append(authorIDs, id)
	}

	query, args := buildInClause(`
        SELECT id, team_name
        FROM users
        WHERE id IN (`, authorIDs)

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("load authors' teams: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// Log error but don't fail - rows are already read
		}
	}()

	authorTeam := make(map[string]string)
	for rows.Next() {
		var (
			id       string
			teamName string
		)
		if err := rows.Scan(&id, &teamName); err != nil {
			return nil, fmt.Errorf("scan author team: %w", err)
		}
		authorTeam[id] = teamName
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate authors' teams: %w", err)
	}
	return authorTeam, nil
}

func (r *PRRepo) loadCandidates(ctx context.Context, tx *sql.Tx, authorTeam map[string]string) (map[string][]string, error) {
	teamSet := make(map[string]struct{})
	for _, t := range authorTeam {
		teamSet[t] = struct{}{}
	}
	teamNames := make([]string, 0, len(teamSet))
	for t := range teamSet {
		teamNames = append(teamNames, t)
	}

	if len(teamNames) == 0 {
		teamNames = nil
	}

	query, args := buildInClause(`
        SELECT id, team_name
        FROM users
        WHERE is_active = true AND team_name IN (`, teamNames)

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("load active candidates: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// Log error but don't fail - rows are already read
		}
	}()

	candidatesByTeam := make(map[string][]string)
	for rows.Next() {
		var (
			id       string
			teamName string
		)
		if err := rows.Scan(&id, &teamName); err != nil {
			return nil, fmt.Errorf("scan candidate: %w", err)
		}
		candidatesByTeam[teamName] = append(candidatesByTeam[teamName], id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate candidates: %w", err)
	}
	return candidatesByTeam, nil
}

func (r *PRRepo) calculateNewReviewers(prMap map[string]*prInfo, authorTeam map[string]string, candidatesByTeam map[string][]string) map[string][]string {
	newReviewersByPR := make(map[string][]string, len(prMap))

	for prID, info := range prMap {
		authorID := info.authorID
		teamName := authorTeam[authorID]

		deactSet := info.deactivated
		present := make(map[string]struct{})
		newReviewers := make([]string, 0, 2)

		for _, id := range info.current {
			if _, gone := deactSet[id]; gone {
				continue
			}
			if _, ok := present[id]; ok {
				continue
			}
			newReviewers = append(newReviewers, id)
			present[id] = struct{}{}
		}

		candidates := candidatesByTeam[teamName]
		for _, cand := range candidates {
			if len(newReviewers) >= 2 {
				break
			}
			if cand == authorID {
				continue
			}
			if _, ok := present[cand]; ok {
				continue
			}
			newReviewers = append(newReviewers, cand)
			present[cand] = struct{}{}
		}

		newReviewersByPR[prID] = newReviewers
	}

	return newReviewersByPR
}

func (r *PRRepo) updatePRReviewers(ctx context.Context, tx *sql.Tx, newReviewersByPR map[string][]string) error {
	for prID, reviewers := range newReviewersByPR {
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM pull_request_reviewers WHERE pr_id = $1`,
			prID,
		); err != nil {
			return fmt.Errorf("delete old reviewers for pr %s: %w", prID, err)
		}

		for _, reviewerID := range reviewers {
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO pull_request_reviewers (pr_id, reviewer_id)
                 VALUES ($1, $2)`,
				prID, reviewerID,
			); err != nil {
				return fmt.Errorf("insert new reviewer %s for pr %s: %w", reviewerID, prID, err)
			}
		}
	}
	return nil
}

func buildInClause(prefix string, ids []string) (string, []any) {
	if len(ids) == 0 {
		return prefix + "NULL)", nil
	}

	query := prefix
	args := make([]any, 0, len(ids))
	for i, id := range ids {
		if i > 0 {
			query += ","
		}
		query += fmt.Sprintf("$%d", i+1)
		args = append(args, id)
	}
	query += ")"
	return query, args
}
