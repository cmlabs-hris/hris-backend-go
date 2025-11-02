package postgresql

import (
	"context"
	"fmt"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/master/position"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/jackc/pgx/v5"
)

type positionRepositoryImpl struct {
	db *database.DB
}

func NewPositionRepository(db *database.DB) position.PositionRepository {
	return &positionRepositoryImpl{db: db}
}

// Create implements position.PositionRepository.
func (r *positionRepositoryImpl) Create(ctx context.Context, p position.Position) (position.Position, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		INSERT INTO positions (id, company_id, name, created_at, updated_at)
		VALUES (uuidv7(), $1, $2, NOW(), NOW())
		RETURNING id, company_id, name
	`

	var result position.Position
	err := q.QueryRow(ctx, query, p.CompanyID, p.Name).Scan(
		&result.ID,
		&result.CompanyID,
		&result.Name,
	)

	if err != nil {
		return position.Position{}, fmt.Errorf("failed to create position: %w", err)
	}

	return result, nil
}

// GetByID implements position.PositionRepository.
func (r *positionRepositoryImpl) GetByID(ctx context.Context, id string) (position.Position, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, company_id, name
		FROM positions
		WHERE id = $1
	`

	var result position.Position
	err := q.QueryRow(ctx, query, id).Scan(
		&result.ID,
		&result.CompanyID,
		&result.Name,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return position.Position{}, fmt.Errorf("position not found")
		}
		return position.Position{}, fmt.Errorf("failed to get position: %w", err)
	}

	return result, nil
}

// GetByCompanyID implements position.PositionRepository.
func (r *positionRepositoryImpl) GetByCompanyID(ctx context.Context, companyID string) ([]position.Position, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, company_id, name
		FROM positions
		WHERE company_id = $1
		ORDER BY name ASC
	`

	rows, err := q.Query(ctx, query, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}
	defer rows.Close()

	var positions []position.Position
	for rows.Next() {
		var p position.Position
		err := rows.Scan(
			&p.ID,
			&p.CompanyID,
			&p.Name,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan position: %w", err)
		}
		positions = append(positions, p)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return positions, nil
}

// Update implements position.PositionRepository.
func (r *positionRepositoryImpl) Update(ctx context.Context, req position.UpdatePositionRequest) error {
	q := GetQuerier(ctx, r.db)

	query := `
		UPDATE positions 
		SET name = $1, updated_at = NOW()
		WHERE id = $2
	`

	commandTag, err := q.Exec(ctx, query, req.Name, req.ID)
	if err != nil {
		return fmt.Errorf("failed to update position: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("position not found")
	}

	return nil
}

// Delete implements position.PositionRepository.
func (r *positionRepositoryImpl) Delete(ctx context.Context, id string) error {
	q := GetQuerier(ctx, r.db)

	query := `DELETE FROM positions WHERE id = $1`

	commandTag, err := q.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete position: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("position not found")
	}

	return nil
}
