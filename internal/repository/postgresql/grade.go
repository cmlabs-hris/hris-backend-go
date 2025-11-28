package postgresql

import (
	"context"
	"fmt"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/master/grade"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/jackc/pgx/v5"
)

type gradeRepositoryImpl struct {
	db *database.DB
}

func NewGradeRepository(db *database.DB) grade.GradeRepository {
	return &gradeRepositoryImpl{db: db}
}

// Create implements grade.GradeRepository.
func (r *gradeRepositoryImpl) Create(ctx context.Context, g grade.Grade) (grade.Grade, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		INSERT INTO grades (id, company_id, name)
		VALUES (uuidv7(), $1, $2)
		RETURNING id, company_id, name
	`

	var result grade.Grade
	err := q.QueryRow(ctx, query, g.CompanyID, g.Name).Scan(
		&result.ID,
		&result.CompanyID,
		&result.Name,
	)

	if err != nil {
		return grade.Grade{}, fmt.Errorf("failed to create grade: %w", err)
	}

	return result, nil
}

// GetByID implements grade.GradeRepository.
func (r *gradeRepositoryImpl) GetByID(ctx context.Context, id string, companyID string) (grade.Grade, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, company_id, name
		FROM grades
		WHERE id = $1 AND company_id = $2
	`

	var result grade.Grade
	err := q.QueryRow(ctx, query, id, companyID).Scan(
		&result.ID,
		&result.CompanyID,
		&result.Name,
	)

	if err == pgx.ErrNoRows {
		return grade.Grade{}, fmt.Errorf("grade not found: %w", err)
	}

	if err != nil {
		return grade.Grade{}, fmt.Errorf("failed to get grade: %w", err)
	}

	return result, nil
}

// GetByCompanyID implements grade.GradeRepository.
func (r *gradeRepositoryImpl) GetByCompanyID(ctx context.Context, companyID string) ([]grade.Grade, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, company_id, name
		FROM grades
		WHERE company_id = $1
		ORDER BY name ASC
	`

	rows, err := q.Query(ctx, query, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get grades: %w", err)
	}
	defer rows.Close()

	var grades []grade.Grade
	for rows.Next() {
		var g grade.Grade
		err := rows.Scan(
			&g.ID,
			&g.CompanyID,
			&g.Name,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan grade: %w", err)
		}
		grades = append(grades, g)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return grades, nil
}

// Update implements grade.GradeRepository.
func (r *gradeRepositoryImpl) Update(ctx context.Context, req grade.UpdateGradeRequest) error {
	q := GetQuerier(ctx, r.db)

	query := `
		UPDATE grades 
		SET name = $1
		WHERE id = $2 AND company_id = $3
	`

	commandTag, err := q.Exec(ctx, query, req.Name, req.ID, req.CompanyID)
	if err != nil {
		return fmt.Errorf("failed to update grade: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("grade not found: %w", pgx.ErrNoRows)
	}

	return nil
}

// Delete implements grade.GradeRepository.
func (r *gradeRepositoryImpl) Delete(ctx context.Context, id string, companyID string) error {
	q := GetQuerier(ctx, r.db)

	query := `DELETE FROM grades WHERE id = $1 AND company_id = $2`

	commandTag, err := q.Exec(ctx, query, id, companyID)
	if err != nil {
		return fmt.Errorf("failed to delete grade: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("grade not found: %w", pgx.ErrNoRows)
	}

	return nil
}
