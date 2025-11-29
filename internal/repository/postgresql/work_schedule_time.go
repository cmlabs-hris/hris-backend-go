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

type workScheduleTimeRepositoryImpl struct {
	db *database.DB
}

// GetByID implements schedule.WorkScheduleTimeRepository.
func (r *workScheduleTimeRepositoryImpl) GetByID(ctx context.Context, id string, companyID string) (schedule.WorkScheduleTime, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, work_schedule_id, day_of_week, clock_in_time, break_start_time,
			   break_end_time, clock_out_time, is_next_day_checkout, location_type, created_at, updated_at
		FROM work_schedule_times
		WHERE id = $1 AND work_schedule_id IN (
			SELECT id FROM work_schedules WHERE company_id = $2
		)
	`

	var t schedule.WorkScheduleTime
	err := q.QueryRow(ctx, query, id, companyID).Scan(
		&t.ID, &t.WorkScheduleID, &t.DayOfWeek, &t.ClockInTime,
		&t.BreakStartTime, &t.BreakEndTime, &t.ClockOutTime, &t.IsNextDayCheckout, &t.LocationType, &t.CreatedAt, &t.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return schedule.WorkScheduleTime{}, fmt.Errorf("work schedule time not found: %w", err)
	}

	if err != nil {
		return schedule.WorkScheduleTime{}, fmt.Errorf("failed to get work schedule time: %w", err)
	}

	return t, nil
}

// Update implements schedule.WorkScheduleTimeRepository.
func (r *workScheduleTimeRepositoryImpl) Update(ctx context.Context, req schedule.UpdateWorkScheduleTimeRequest) error {
	q := GetQuerier(ctx, r.db)

	updates := []string{
		"clock_in_time = $1",
		"clock_out_time = $2",
		"is_next_day_checkout = $3",
		"location_type = $4",
		"updated_at = $5",
	}
	args := []interface{}{
		req.ClockInTime,
		req.ClockOutTime,
		*req.IsNextDayCheckout,
		req.LocationType,
		time.Now(),
	}
	argIdx := 6

	// Optional fields
	if req.DayOfWeek != nil {
		updates = append(updates, fmt.Sprintf("day_of_week = $%d", argIdx))
		args = append(args, *req.DayOfWeek)
		argIdx++
	}
	if req.BreakStartTime != nil {
		updates = append(updates, fmt.Sprintf("break_start_time = $%d", argIdx))
		args = append(args, *req.BreakStartTime)
		argIdx++
	} else {
		updates = append(updates, fmt.Sprintf("break_start_time = $%d", argIdx))
		args = append(args, nil)
		argIdx++
	}
	if req.BreakEndTime != nil {
		updates = append(updates, fmt.Sprintf("break_end_time = $%d", argIdx))
		args = append(args, *req.BreakEndTime)
		argIdx++
	} else {
		updates = append(updates, fmt.Sprintf("break_end_time = $%d", argIdx))
		args = append(args, nil)
		argIdx++
	}

	args = append(args, req.ID)
	idIdx := argIdx
	argIdx++

	args = append(args, req.CompanyID)

	sql := "UPDATE work_schedule_times SET " + strings.Join(updates, ", ") +
		fmt.Sprintf(" WHERE id = $%d AND EXISTS (SELECT 1 FROM work_schedules WHERE id = work_schedule_times.work_schedule_id AND company_id = $%d) RETURNING id", idIdx, argIdx)

	var updatedID string
	if err := q.QueryRow(ctx, sql, args...).Scan(&updatedID); err != nil {
		if err == pgx.ErrNoRows {
			return schedule.ErrWorkScheduleTimeNotFound
		}
		return fmt.Errorf("failed to update work schedule time: %w", err)
	}
	return nil
}

// Create implements schedule.WorkScheduleTimeRepository.
func (r *workScheduleTimeRepositoryImpl) Create(ctx context.Context, time schedule.WorkScheduleTime, companyID string) (schedule.WorkScheduleTime, error) {
	q := GetQuerier(ctx, r.db)

	// Verify work_schedule belongs to company before inserting
	query := `
		INSERT INTO work_schedule_times (
			work_schedule_id, day_of_week, clock_in_time, break_start_time,
			break_end_time, clock_out_time, is_next_day_checkout, location_type
		)
		SELECT $1, $2, $3, $4, $5, $6, $7, $8
		WHERE EXISTS (
			SELECT 1 FROM work_schedules
			WHERE id = $1 AND company_id = $9 AND deleted_at IS NULL
		)
		RETURNING id
	`

	err := q.QueryRow(ctx, query,
		time.WorkScheduleID, time.DayOfWeek, time.ClockInTime,
		time.BreakStartTime, time.BreakEndTime, time.ClockOutTime, time.IsNextDayCheckout, time.LocationType,
		companyID,
	).Scan(&time.ID)

	if err != nil {
		if err == pgx.ErrNoRows {
			return schedule.WorkScheduleTime{}, schedule.ErrWorkScheduleNotFound
		}
		return schedule.WorkScheduleTime{}, fmt.Errorf("failed to create work schedule time: %w", err)
	}

	return time, nil
}

// GetByWorkScheduleID implements schedule.WorkScheduleTimeRepository.
func (r *workScheduleTimeRepositoryImpl) GetByWorkScheduleID(ctx context.Context, scheduleID, companyID string) ([]schedule.WorkScheduleTime, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT wst.id, wst.work_schedule_id, wst.day_of_week, wst.clock_in_time, wst.break_start_time,
			   wst.break_end_time, wst.clock_out_time, wst.is_next_day_checkout, wst.location_type, wst.created_at, wst.updated_at
		FROM work_schedule_times wst
		JOIN work_schedules ws ON wst.work_schedule_id = ws.id
		WHERE wst.work_schedule_id = $1 AND ws.company_id = $2 AND ws.deleted_at IS NULL
		ORDER BY wst.day_of_week
	`

	rows, err := q.Query(ctx, query, scheduleID, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get work schedule times: %w", err)
	}
	defer rows.Close()

	var times []schedule.WorkScheduleTime
	for rows.Next() {
		var t schedule.WorkScheduleTime
		err := rows.Scan(
			&t.ID, &t.WorkScheduleID, &t.DayOfWeek, &t.ClockInTime,
			&t.BreakStartTime, &t.BreakEndTime, &t.ClockOutTime, &t.IsNextDayCheckout, &t.LocationType,
			&t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan work schedule time: %w", err)
		}
		times = append(times, t)
	}

	return times, nil
}

// GetTimeByScheduleAndDay implements schedule.WorkScheduleTimeRepository.
func (r *workScheduleTimeRepositoryImpl) GetTimeByScheduleAndDay(ctx context.Context, scheduleID string, dayOfWeek int, companyID string) (schedule.WorkScheduleTime, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT wst.id, wst.work_schedule_id, wst.day_of_week, wst.clock_in_time, wst.break_start_time,
			   wst.break_end_time, wst.clock_out_time, wst.is_next_day_checkout, wst.location_type, wst.created_at, wst.updated_at
		FROM work_schedule_times wst
		JOIN work_schedules ws ON wst.work_schedule_id = ws.id
		WHERE wst.work_schedule_id = $1 AND wst.day_of_week = $2 AND ws.company_id = $3 AND ws.deleted_at IS NULL
	`

	var t schedule.WorkScheduleTime
	err := q.QueryRow(ctx, query, scheduleID, dayOfWeek, companyID).Scan(
		&t.ID, &t.WorkScheduleID, &t.DayOfWeek, &t.ClockInTime,
		&t.BreakStartTime, &t.BreakEndTime, &t.ClockOutTime, &t.IsNextDayCheckout, &t.LocationType,
		&t.CreatedAt, &t.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return schedule.WorkScheduleTime{}, fmt.Errorf("work schedule time not found: %w", err)
		}
		return schedule.WorkScheduleTime{}, fmt.Errorf("failed to get work schedule time: %w", err)
	}

	return t, nil
}

// Delete implements schedule.WorkScheduleTimeRepository.
func (r *workScheduleTimeRepositoryImpl) Delete(ctx context.Context, id, companyID string) error {
	q := GetQuerier(ctx, r.db)

	query := `DELETE FROM work_schedule_times 
		WHERE id = $1 AND EXISTS (
			SELECT 1 FROM work_schedules 
			WHERE id = work_schedule_times.work_schedule_id AND company_id = $2
		)`

	commandTag, err := q.Exec(ctx, query, id, companyID)
	if err != nil {
		return fmt.Errorf("failed to delete work schedule time: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return schedule.ErrWorkScheduleTimeNotFound
	}

	return nil
}

func NewWorkScheduleTimeRepository(db *database.DB) schedule.WorkScheduleTimeRepository {
	return &workScheduleTimeRepositoryImpl{db: db}
}
