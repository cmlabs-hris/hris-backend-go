package postgresql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/schedule"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/jackc/pgx/v5"
)

type workScheduleLocationRepository struct {
	db *database.DB
}

// BulkDeleteByWorkScheduleID implements schedule.WorkScheduleLocationRepository.
func (w *workScheduleLocationRepository) BulkDeleteByWorkScheduleID(ctx context.Context, workScheduleID string, companyID string) error {
	q := GetQuerier(ctx, w.db)

	query := `DELETE FROM work_schedule_locations 
		WHERE work_schedule_id = $1 AND EXISTS (
			SELECT 1 FROM work_schedules 
			WHERE id = $1 AND company_id = $2
		)`

	_, err := q.Exec(ctx, query, workScheduleID, companyID)
	if err != nil {
		return fmt.Errorf("failed to bulk delete work schedule locations: %w", err)
	}

	return nil
}

// Create implements schedule.WorkScheduleLocationRepository.
func (w *workScheduleLocationRepository) Create(ctx context.Context, workScheduleLocation schedule.WorkScheduleLocation, companyID string) (schedule.WorkScheduleLocation, error) {
	q := GetQuerier(ctx, w.db)

	// Verify work_schedule belongs to company before inserting
	query := `
		INSERT INTO work_schedule_locations (
			work_schedule_id, location_name, latitude, longitude, radius_meters
		)
		SELECT $1, $2, $3, $4, $5
		WHERE EXISTS (
			SELECT 1 FROM work_schedules
			WHERE id = $1 AND company_id = $6 AND deleted_at IS NULL
		)
		RETURNING id
	`

	err := q.QueryRow(ctx, query,
		workScheduleLocation.WorkScheduleID, workScheduleLocation.LocationName, workScheduleLocation.Latitude,
		workScheduleLocation.Longitude, workScheduleLocation.RadiusMeters,
		companyID,
	).Scan(&workScheduleLocation.ID)

	if err != nil {
		if err == pgx.ErrNoRows {
			return schedule.WorkScheduleLocation{}, schedule.ErrWorkScheduleNotFound
		}
		return schedule.WorkScheduleLocation{}, fmt.Errorf("failed to create work schedule location: %w", err)
	}

	return workScheduleLocation, nil
}

// Delete implements schedule.WorkScheduleLocationRepository.
func (w *workScheduleLocationRepository) Delete(ctx context.Context, id, companyID string) error {
	q := GetQuerier(ctx, w.db)

	query := `DELETE FROM work_schedule_locations 
		WHERE id = $1 AND EXISTS (
			SELECT 1 FROM work_schedules 
			WHERE id = work_schedule_locations.work_schedule_id AND company_id = $2
		)`

	commandTag, err := q.Exec(ctx, query, id, companyID)
	if err != nil {
		return fmt.Errorf("failed to delete work schedule location: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return schedule.ErrWorkScheduleLocationNotFound
	}

	return nil
}

// GetByID implements schedule.WorkScheduleLocationRepository.
func (w *workScheduleLocationRepository) GetByID(ctx context.Context, id string, companyID string) (schedule.WorkScheduleLocation, error) {
	q := GetQuerier(ctx, w.db)

	query := `
		SELECT wsl.id, wsl.work_schedule_id, wsl.location_name, wsl.latitude, wsl.longitude, wsl.radius_meters, wsl.created_at, wsl.updated_at
		FROM work_schedule_locations wsl
		JOIN work_schedules ws ON wsl.work_schedule_id = ws.id
		WHERE wsl.id = $1 AND ws.company_id = $2
	`

	var location schedule.WorkScheduleLocation
	err := q.QueryRow(ctx, query, id, companyID).Scan(
		&location.ID, &location.WorkScheduleID, &location.LocationName,
		&location.Latitude, &location.Longitude, &location.RadiusMeters,
		&location.CreatedAt, &location.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return schedule.WorkScheduleLocation{}, fmt.Errorf("work schedule location not found: %w", err)
		}
		return schedule.WorkScheduleLocation{}, fmt.Errorf("failed to get work schedule location: %w", err)
	}

	return location, nil
}

// GetByWorkScheduleID implements schedule.WorkScheduleLocationRepository.
func (w *workScheduleLocationRepository) GetByWorkScheduleID(ctx context.Context, workScheduleID, companyID string) ([]schedule.WorkScheduleLocation, error) {
	q := GetQuerier(ctx, w.db)

	query := `
		SELECT wsl.id, wsl.work_schedule_id, wsl.location_name, wsl.latitude, wsl.longitude, wsl.radius_meters, wsl.created_at, wsl.updated_at
		FROM work_schedule_locations wsl
		JOIN work_schedules ws ON wsl.work_schedule_id = ws.id
		WHERE wsl.work_schedule_id = $1 AND ws.company_id = $2 AND ws.deleted_at IS NULL
	`

	rows, err := q.Query(ctx, query, workScheduleID, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get work schedule locations: %w", err)
	}
	defer rows.Close()

	var locations []schedule.WorkScheduleLocation
	for rows.Next() {
		var loc schedule.WorkScheduleLocation
		err := rows.Scan(
			&loc.ID, &loc.WorkScheduleID, &loc.LocationName, &loc.Latitude, &loc.Longitude, &loc.RadiusMeters,
			&loc.CreatedAt, &loc.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan work schedule location: %w", err)
		}
		locations = append(locations, loc)
	}

	return locations, nil
}

// Update implements schedule.WorkScheduleLocationRepository.
func (w *workScheduleLocationRepository) Update(ctx context.Context, req schedule.UpdateWorkScheduleLocationRequest) error {
	q := GetQuerier(ctx, w.db)

	updates := make([]string, 0)
	args := make([]interface{}, 0)
	argIdx := 1

	if req.LocationName != nil {
		updates = append(updates, fmt.Sprintf("location_name = $%d", argIdx))
		args = append(args, *req.LocationName)
		argIdx++
	}
	if req.Latitude != nil {
		updates = append(updates, fmt.Sprintf("latitude = $%d", argIdx))
		args = append(args, *req.Latitude)
		argIdx++
	}
	if req.Longitude != nil {
		updates = append(updates, fmt.Sprintf("longitude = $%d", argIdx))
		args = append(args, *req.Longitude)
		argIdx++
	}
	if req.RadiusMeters != nil {
		updates = append(updates, fmt.Sprintf("radius_meters = $%d", argIdx))
		args = append(args, *req.RadiusMeters)
		argIdx++
	}

	if len(updates) == 0 {
		return fmt.Errorf("no updatable fields provided for work schedule location update")
	}

	updates = append(updates, fmt.Sprintf("updated_at = $%d", argIdx))
	args = append(args, time.Now())
	argIdx++

	args = append(args, req.ID)
	idIdx := argIdx
	argIdx++

	args = append(args, req.CompanyID)

	sql := "UPDATE work_schedule_locations SET " + strings.Join(updates, ", ") +
		fmt.Sprintf(" WHERE id = $%d AND EXISTS (SELECT 1 FROM work_schedules WHERE id = work_schedule_locations.work_schedule_id AND company_id = $%d) RETURNING id", idIdx, argIdx)

	var updatedID string
	if err := q.QueryRow(ctx, sql, args...).Scan(&updatedID); err != nil {
		if err == pgx.ErrNoRows {
			return schedule.ErrWorkScheduleLocationNotFound
		}
		return fmt.Errorf("failed to update work schedule location: %w", err)
	}
	return nil
}

func NewWorkScheduleLocationRepository(db *database.DB) schedule.WorkScheduleLocationRepository {
	return &workScheduleLocationRepository{db: db}
}
