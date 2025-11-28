package postgresql

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/schedule"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/jackc/pgx/v5"
)

type workScheduleRepositoryImpl struct {
	db *database.DB
}

// GetActiveSchedule implements schedule.WorkScheduleRepository.
func (w *workScheduleRepositoryImpl) GetActiveSchedule(ctx context.Context, employeeID string, date time.Time, companyID string) (*schedule.ActiveSchedule, error) {
	q := GetQuerier(ctx, w.db)

	query := `
		-- QUERY: GetActiveSchedule

WITH target_schedule AS (
    SELECT COALESCE(
        -- PRIORITAS 1: Cek Override (Assignments)
        (
            SELECT work_schedule_id 
            FROM employee_schedule_assignments 
            WHERE employee_id = $1 
              AND $2::date BETWEEN start_date AND end_date
            LIMIT 1
        ),
        -- PRIORITAS 2: Cek Default (Employee Master)
        (
            SELECT work_schedule_id 
            FROM employees 
            WHERE id = $1 AND company_id = $3
        )
    ) AS id
)
SELECT 
    ws.id AS schedule_id,
    ws.name AS schedule_name,
    ws.grace_period_minutes,
    ws.type AS location_type, -- 'WFO', 'WFA', 'Hybrid'
    
    -- Detail Waktu (Spesifik Hari Ini)
    wst.id AS time_id,
    wst.clock_in_time,
    wst.clock_out_time,
    wst.is_next_day_checkout,
    
    -- Detail Lokasi (Digabung jadi satu JSON Array)
    COALESCE(
        (
            SELECT json_agg(json_build_object(
                'name', wsl.location_name,
                'latitude', wsl.latitude,
                'longitude', wsl.longitude,
                'radius_meters', wsl.radius_meters
            ))
            FROM work_schedule_locations wsl
            WHERE wsl.work_schedule_id = ws.id
        ), 
        '[]'::json
    ) AS allowed_locations

FROM target_schedule ts
JOIN work_schedules ws ON ws.id = ts.id
-- JOIN PENTING: Hanya ambil aturan untuk HARI yang diminta
-- EXTRACT(ISODOW) mengembalikan 1 (Senin) s/d 7 (Minggu)
JOIN work_schedule_times wst ON wst.work_schedule_id = ws.id 
    AND wst.day_of_week = EXTRACT(ISODOW FROM $2::date)::int

WHERE 
    ws.company_id = $3
    AND ws.deleted_at IS NULL; -- Pastikan jadwal belum di-soft delete
	`

	// ActiveScheduleDTO menampung hasil raw dari database
	type activeScheduleDTO struct {
		ScheduleID         string `db:"schedule_id"`
		ScheduleName       string `db:"schedule_name"`
		LocationType       string `db:"location_type"`
		GracePeriodMinutes int    `db:"grace_period_minutes"`

		// Detail Waktu
		TimeID            string    `db:"time_id"`
		ClockInTime       time.Time `db:"clock_in_time"`
		ClockOutTime      time.Time `db:"clock_out_time"`
		IsNextDayCheckout bool      `db:"is_next_day_checkout"`

		// Lokasi (Raw JSON)
		AllowedLocationsJSON []byte `db:"allowed_locations"`
	}

	var dto activeScheduleDTO

	err := q.QueryRow(ctx, query, employeeID, date, companyID).Scan(
		&dto.ScheduleID,
		&dto.ScheduleName,
		&dto.GracePeriodMinutes,
		&dto.LocationType,
		&dto.TimeID,
		&dto.ClockInTime,
		&dto.ClockOutTime,
		&dto.IsNextDayCheckout,
		&dto.AllowedLocationsJSON,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get active schedule: %w", err)
	}

	// Parse JSON locations
	var locations []schedule.ScheduleLocation
	if err := json.Unmarshal(dto.AllowedLocationsJSON, &locations); err != nil {
		return nil, fmt.Errorf("failed to parse locations: %w", err)
	}

	// Map DTO to domain model
	return &schedule.ActiveSchedule{
		ScheduleID:         dto.ScheduleID,
		ScheduleName:       dto.ScheduleName,
		LocationType:       dto.LocationType,
		GracePeriodMinutes: dto.GracePeriodMinutes,
		TimeID:             dto.TimeID,
		ClockIn:            dto.ClockInTime,
		ClockOut:           dto.ClockOutTime,
		IsNextDayCheckout:  dto.IsNextDayCheckout,
		Locations:          locations,
	}, nil
}

// Create implements schedule.WorkScheduleRepository.
func (w *workScheduleRepositoryImpl) Create(ctx context.Context, workSchedule schedule.WorkSchedule) (schedule.WorkSchedule, error) {
	q := GetQuerier(ctx, w.db)

	query := `
		INSERT INTO work_schedules (
			id, company_id, name, type, created_at, updated_at
		) VALUES (
			uuidv7(), $1, $2, $3, NOW(), NOW()
		) RETURNING id, created_at, updated_at
	`

	err := q.QueryRow(ctx, query,
		workSchedule.CompanyID, workSchedule.Name, workSchedule.Type,
	).Scan(&workSchedule.ID, &workSchedule.CreatedAt, &workSchedule.UpdatedAt)

	if err != nil {
		return schedule.WorkSchedule{}, err
	}

	return workSchedule, nil
}

// Delete implements schedule.WorkScheduleRepository.
func (w *workScheduleRepositoryImpl) Delete(ctx context.Context, id, companyID string) error {
	q := GetQuerier(ctx, w.db)
	query := `
		DELETE FROM work_schedules
		WHERE id = $1 AND company_id = $2
	`
	commandTag, err := q.Exec(ctx, query, id, companyID)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return schedule.ErrWorkScheduleNotFound
	}
	return nil
}

// GetByCompanyID implements schedule.WorkScheduleRepository.
func (w *workScheduleRepositoryImpl) GetByCompanyID(ctx context.Context, companyID string, filter schedule.WorkScheduleFilter) ([]schedule.WorkSchedule, int64, error) {
	q := GetQuerier(ctx, w.db)

	// Base query
	baseQuery := `
		FROM work_schedules
		WHERE company_id = $1 AND deleted_at IS NULL
	`

	args := []interface{}{companyID}
	argIdx := 2

	// Build WHERE clause dynamically
	whereClauses := []string{}

	// Filter by name (ILIKE for case-insensitive search)
	if filter.Name != nil && *filter.Name != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("name ILIKE $%d", argIdx))
		args = append(args, "%"+*filter.Name+"%")
		argIdx++
	}

	// Filter by type
	if filter.Type != nil && *filter.Type != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("type = $%d", argIdx))
		args = append(args, *filter.Type)
		argIdx++
	}

	// Append WHERE clauses
	if len(whereClauses) > 0 {
		baseQuery += " AND " + strings.Join(whereClauses, " AND ")
	}

	// COUNT query for total records
	countQuery := "SELECT COUNT(*) " + baseQuery
	var total int64
	err := q.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count work schedules: %w", err)
	}

	// Main SELECT query
	selectQuery := `
		SELECT id, company_id, name, type, created_at, updated_at
	` + baseQuery

	// ORDER BY clause
	orderBy := "name ASC" // Default
	switch filter.SortBy {
	case "type":
		orderBy = "type"
	case "created_at":
		orderBy = "created_at"
	case "updated_at":
		orderBy = "updated_at"
	}

	if strings.ToLower(filter.SortOrder) == "desc" {
		orderBy += " DESC"
	} else {
		orderBy += " ASC"
	}

	selectQuery += " ORDER BY " + orderBy

	// Pagination
	limit := filter.Limit
	if limit == 0 {
		limit = 20
	}
	offset := (filter.Page - 1) * limit

	selectQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	// Execute query
	rows, err := q.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query work schedules: %w", err)
	}
	defer rows.Close()

	return w.mapRowsToWorkSchedules(rows, total)
}

// GetByID implements schedule.WorkScheduleRepository.
func (w *workScheduleRepositoryImpl) GetByID(ctx context.Context, id string, companyID string) (schedule.WorkSchedule, error) {
	q := GetQuerier(ctx, w.db)
	query := `
		SELECT id, company_id, name, type, created_at, updated_at
		FROM work_schedules
		WHERE id = $1 AND company_id = $2 AND deleted_at IS NULL
	`

	var ws schedule.WorkSchedule
	err := q.QueryRow(ctx, query, id, companyID).Scan(
		&ws.ID, &ws.CompanyID, &ws.Name, &ws.Type, &ws.CreatedAt, &ws.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return schedule.WorkSchedule{}, fmt.Errorf("work schedule not found: %w", err)
		}
		return schedule.WorkSchedule{}, err
	}

	return ws, nil
}

// Update implements schedule.WorkScheduleRepository.
func (w *workScheduleRepositoryImpl) Update(ctx context.Context, req schedule.UpdateWorkScheduleRequest) (schedule.WorkSchedule, error) {
	q := GetQuerier(ctx, w.db)

	updates := make([]string, 0)
	args := make([]interface{}, 0)
	argIdx := 1

	if req.Name != nil {
		updates = append(updates, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *req.Name)
		argIdx++
	}
	if req.Type != nil {
		updates = append(updates, fmt.Sprintf("type = $%d", argIdx))
		args = append(args, *req.Type)
		argIdx++
	}

	if len(updates) == 0 {
		return schedule.WorkSchedule{}, fmt.Errorf("no updatable fields provided for work schedule update")
	}

	updates = append(updates, fmt.Sprintf("updated_at = $%d", argIdx))
	args = append(args, time.Now())
	argIdx++

	args = append(args, req.ID)
	idIdx := argIdx
	argIdx++

	args = append(args, req.CompanyID)

	query := "UPDATE work_schedules SET " + strings.Join(updates, ", ") +
		fmt.Sprintf(" WHERE id = $%d AND company_id = $%d RETURNING id, company_id, name, type, grace_period_minutes, created_at, updated_at", idIdx, argIdx)

	var ws schedule.WorkSchedule
	err := q.QueryRow(ctx, query, args...).Scan(
		&ws.ID, &ws.CompanyID, &ws.Name, &ws.Type, &ws.GracePeriodMinutes, &ws.CreatedAt, &ws.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return schedule.WorkSchedule{}, schedule.ErrWorkScheduleNotFound
		}
		return schedule.WorkSchedule{}, fmt.Errorf("failed to update work schedule: %w", err)
	}

	return ws, nil
}

func (r *workScheduleRepositoryImpl) mapRowsToWorkSchedules(rows pgx.Rows, total int64) ([]schedule.WorkSchedule, int64, error) {
	// Map untuk aggregate hasil JOIN
	schedulesMap := make(map[string]*schedule.WorkSchedule)

	for rows.Next() {
		var raw workScheduleWithRelations

		err := rows.Scan(
			&raw.WorkScheduleID, &raw.CompanyID, &raw.Name, &raw.Type,
			&raw.TimeID, &raw.DayOfWeek, &raw.ClockInTime, &raw.ClockOutTime,
			&raw.BreakStartTime, &raw.BreakEndTime, &raw.LocationType,
			&raw.LocationID, &raw.LocationName, &raw.Latitude, &raw.Longitude, &raw.RadiusMeters,
		)
		if err != nil {
			return nil, total, fmt.Errorf("failed to scan row: %w", err)
		}

		// Get or create WorkSchedule
		ws, exists := schedulesMap[raw.WorkScheduleID]
		if !exists {
			ws = &schedule.WorkSchedule{
				ID:        raw.WorkScheduleID,
				CompanyID: raw.CompanyID,
				Name:      raw.Name,
				Type:      schedule.WorkArrangement(raw.Type),
				Times:     []schedule.WorkScheduleTime{},
				Locations: []schedule.WorkScheduleLocation{},
			}
			schedulesMap[raw.WorkScheduleID] = ws
		}

		// Append Time (jika ada)
		if raw.TimeID != nil {
			// Check duplicate (karena bisa multiple locations untuk same time)
			isDuplicate := false
			for _, t := range ws.Times {
				if t.ID == *raw.TimeID {
					isDuplicate = true
					break
				}
			}

			if !isDuplicate {
				ws.Times = append(ws.Times, schedule.WorkScheduleTime{
					ID:             *raw.TimeID,
					WorkScheduleID: raw.WorkScheduleID,
					DayOfWeek:      *raw.DayOfWeek,
					ClockInTime:    *raw.ClockInTime,
					ClockOutTime:   *raw.ClockOutTime,
					BreakStartTime: raw.BreakStartTime,
					BreakEndTime:   raw.BreakEndTime,
					LocationType:   schedule.WorkArrangement(*raw.LocationType),
				})
			}
		}

		// Append Location (jika ada)
		if raw.LocationID != nil {
			// Check duplicate
			isDuplicate := false
			for _, loc := range ws.Locations {
				if loc.ID == *raw.LocationID {
					isDuplicate = true
					break
				}
			}

			if !isDuplicate {
				ws.Locations = append(ws.Locations, schedule.WorkScheduleLocation{
					ID:             *raw.LocationID,
					WorkScheduleID: raw.WorkScheduleID,
					LocationName:   *raw.LocationName,
					Latitude:       *raw.Latitude,
					Longitude:      *raw.Longitude,
					RadiusMeters:   *raw.RadiusMeters,
				})
			}
		}
	}

	// Convert map to slice
	var result []schedule.WorkSchedule
	for _, ws := range schedulesMap {
		result = append(result, *ws)
	}

	return result, total, nil
}

// SoftDelete implements schedule.WorkScheduleRepository.
func (w *workScheduleRepositoryImpl) SoftDelete(ctx context.Context, id, companyID string) error {
	q := GetQuerier(ctx, w.db)
	query := `
		UPDATE work_schedules
		SET deleted_at = NOW()
		WHERE id = $1 AND company_id = $2 AND deleted_at IS NULL
	`
	commandTag, err := q.Exec(ctx, query, id, companyID)
	if err != nil {
		return fmt.Errorf("failed to soft delete work schedule: %w", err)
	}
	if commandTag.RowsAffected() != 1 {
		return schedule.ErrWorkScheduleNotFound
	}
	return nil
}

// GetEmployeeScheduleTimeline implements schedule.WorkScheduleRepository.
func (w *workScheduleRepositoryImpl) GetEmployeeScheduleTimeline(ctx context.Context, employeeID string, companyID string, page int, limit int) ([]schedule.EmployeeScheduleTimelineItem, int64, string, error) {
	q := GetQuerier(ctx, w.db)

	// Calculate offset
	offset := (page - 1) * limit

	// CTE to get employee info and validate company_id
	// Then UNION ALL to combine default schedule and override schedules
	query := `
		WITH employee_info AS (
			SELECT 
				e.id,
				e.full_name,
				e.work_schedule_id,
				e.hire_date,
				e.company_id
			FROM employees e
			WHERE e.id = $1 
				AND e.company_id = $2
				AND e.deleted_at IS NULL
		),
		default_schedule AS (
			SELECT
				NULL::UUID as assignment_id,
				'default' as type,
				ei.work_schedule_id as schedule_id,
				ws.name as schedule_name,
				ws.type as schedule_type,
				ws.grace_period_minutes,
				ei.hire_date::DATE as start_date,
				NULL::DATE as end_date,
				1 as sort_priority  -- default comes last
			FROM employee_info ei
			LEFT JOIN work_schedules ws ON ws.id = ei.work_schedule_id AND ws.deleted_at IS NULL
			WHERE ei.work_schedule_id IS NOT NULL
		),
		override_schedules AS (
			SELECT
				esa.id as assignment_id,
				'override' as type,
				esa.work_schedule_id as schedule_id,
				ws.name as schedule_name,
				ws.type as schedule_type,
				ws.grace_period_minutes,
				esa.start_date,
				esa.end_date,
				0 as sort_priority  -- overrides come first
			FROM employee_info ei
			JOIN employee_schedule_assignments esa ON esa.employee_id = ei.id
			JOIN work_schedules ws ON ws.id = esa.work_schedule_id AND ws.deleted_at IS NULL
		),
		combined AS (
			SELECT * FROM override_schedules
			UNION ALL
			SELECT * FROM default_schedule
		),
		total_count AS (
			SELECT COUNT(*) as total FROM combined
		)
		SELECT 
			c.assignment_id,
			c.type,
			c.schedule_id,
			c.schedule_name,
			c.schedule_type,
			c.grace_period_minutes,
			c.start_date,
			c.end_date,
			tc.total,
			ei.full_name
		FROM combined c
		CROSS JOIN total_count tc
		CROSS JOIN employee_info ei
		ORDER BY c.sort_priority ASC, c.start_date DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := q.Query(ctx, query, employeeID, companyID, limit, offset)
	if err != nil {
		return nil, 0, "", fmt.Errorf("failed to get employee schedule timeline: %w", err)
	}
	defer rows.Close()

	var items []schedule.EmployeeScheduleTimelineItem
	var total int64
	var employeeName string

	for rows.Next() {
		var item schedule.EmployeeScheduleTimelineItem
		var assignmentID *string
		var scheduleID string
		var scheduleName string
		var scheduleType string
		var gracePeriodMinutes int
		var startDate *time.Time
		var endDate *time.Time

		err := rows.Scan(
			&assignmentID,
			&item.Type,
			&scheduleID,
			&scheduleName,
			&scheduleType,
			&gracePeriodMinutes,
			&startDate,
			&endDate,
			&total,
			&employeeName,
		)
		if err != nil {
			return nil, 0, "", fmt.Errorf("failed to scan timeline item: %w", err)
		}

		item.ID = assignmentID
		item.ScheduleSnapshot.ID = scheduleID
		item.ScheduleSnapshot.Name = scheduleName
		item.ScheduleSnapshot.Type = scheduleType
		item.ScheduleSnapshot.GracePeriodMinutes = gracePeriodMinutes

		// Convert dates to string pointers
		if startDate != nil {
			dateStr := startDate.Format("2006-01-02")
			item.DateRange.Start = &dateStr
		}
		if endDate != nil {
			dateStr := endDate.Format("2006-01-02")
			item.DateRange.End = &dateStr
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, "", fmt.Errorf("error iterating timeline rows: %w", err)
	}

	// If no rows found at all, return error
	if len(items) == 0 {
		return nil, 0, "", pgx.ErrNoRows
	}

	return items, total, employeeName, nil
}

func NewWorkScheduleRepository(db *database.DB) schedule.WorkScheduleRepository {
	return &workScheduleRepositoryImpl{db: db}
}

// Internal DTO for query result (tidak expose ke domain)
type workScheduleWithRelations struct {
	// Work Schedule fields
	WorkScheduleID string
	CompanyID      string
	Name           string
	Type           string

	// Work Schedule Time fields (nullable karena LEFT JOIN)
	TimeID         *string
	DayOfWeek      *int
	ClockInTime    *time.Time
	ClockOutTime   *time.Time
	BreakStartTime *time.Time
	BreakEndTime   *time.Time
	LocationType   *string

	// Work Schedule Location fields (nullable karena LEFT JOIN)
	LocationID   *string
	LocationName *string
	Latitude     *float64
	Longitude    *float64
	RadiusMeters *int
}
