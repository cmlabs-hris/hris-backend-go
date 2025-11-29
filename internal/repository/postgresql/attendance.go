package postgresql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/attendance"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/jackc/pgx/v5"
)

type attendanceRepository struct {
	db *database.DB
}

// GetOpenSession implements attendance.AttendanceRepository.
func (a *attendanceRepository) GetOpenSession(ctx context.Context, employeeID string) (attendance.Attendance, error) {
	q := GetQuerier(ctx, a.db)

	query := `
		SELECT id, employee_id, company_id, date, work_schedule_time_id, actual_location_type,
			   clock_in, clock_out, work_hours_in_minutes,
			   clock_in_latitude, clock_in_longitude, clock_in_proof_url,
			   clock_out_latitude, clock_out_longitude, clock_out_proof_url,
			   status, approved_by, approved_at, rejection_reason,
			   leave_type_id, late_minutes, early_leave_minutes, overtime_minutes,
			   created_at, updated_at
		FROM attendances
		WHERE employee_id = $1
		  AND clock_out IS NULL
		ORDER BY clock_in DESC
		LIMIT 1
	`

	var att attendance.Attendance
	err := q.QueryRow(ctx, query, employeeID).Scan(
		&att.ID, &att.EmployeeID, &att.CompanyID, &att.Date, &att.WorkScheduleTimeID, &att.ActualLocationType,
		&att.ClockIn, &att.ClockOut, &att.WorkHoursInMinutes,
		&att.ClockInLatitude, &att.ClockInLongitude, &att.ClockInProofURL,
		&att.ClockOutLatitude, &att.ClockOutLongitude, &att.ClockOutProofURL,
		&att.Status, &att.ApprovedBy, &att.ApprovedAt, &att.RejectionReason,
		&att.LeaveTypeID, &att.LateMinutes, &att.EarlyLeaveMinutes, &att.OvertimeMinutes,
		&att.CreatedAt, &att.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return attendance.Attendance{}, fmt.Errorf("no open attendance session found: %w", err)
		}
		return attendance.Attendance{}, fmt.Errorf("failed to get open session: %w", err)
	}

	return att, nil
}

// Create implements attendance.AttendanceRepository.
func (a *attendanceRepository) Create(ctx context.Context, newAttendance attendance.Attendance) (attendance.Attendance, error) {
	q := GetQuerier(ctx, a.db)

	query := `
		INSERT INTO attendances (
			employee_id, company_id, date, work_schedule_time_id, actual_location_type,
			clock_in, clock_in_latitude, clock_in_longitude, clock_in_proof_url,
			status, late_minutes, early_leave_minutes, overtime_minutes, leave_type_id,
			approved_by, approved_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		) RETURNING id, created_at, updated_at
	`

	err := q.QueryRow(ctx, query,
		newAttendance.EmployeeID,
		newAttendance.CompanyID,
		newAttendance.Date,
		newAttendance.WorkScheduleTimeID,
		newAttendance.ActualLocationType,
		newAttendance.ClockIn,
		newAttendance.ClockInLatitude,
		newAttendance.ClockInLongitude,
		newAttendance.ClockInProofURL,
		newAttendance.Status,
		newAttendance.LateMinutes,
		newAttendance.EarlyLeaveMinutes,
		newAttendance.OvertimeMinutes,
		newAttendance.LeaveTypeID,
		newAttendance.ApprovedBy,
		newAttendance.ApprovedAt,
	).Scan(&newAttendance.ID, &newAttendance.CreatedAt, &newAttendance.UpdatedAt)

	if err != nil {
		return attendance.Attendance{}, fmt.Errorf("failed to create attendance: %w", err)
	}

	return newAttendance, nil
}

// GetByEmployeeAndDate implements attendance.AttendanceRepository.
func (a *attendanceRepository) GetByEmployeeAndDate(ctx context.Context, employeeID string, date time.Time, companyID string) (*attendance.Attendance, error) {
	q := GetQuerier(ctx, a.db)

	query := `
		SELECT id, employee_id, company_id, date, work_schedule_time_id, actual_location_type,
			   clock_in, clock_out, work_hours_in_minutes,
			   clock_in_latitude, clock_in_longitude, clock_in_proof_url,
			   clock_out_latitude, clock_out_longitude, clock_out_proof_url,
			   status, approved_by, approved_at, rejection_reason,
			   leave_type_id, late_minutes, early_leave_minutes, overtime_minutes,
			   created_at, updated_at
		FROM attendances
		WHERE employee_id = $1
		  AND date = $2
		  AND company_id = $3
		LIMIT 1
	`

	var att attendance.Attendance
	err := q.QueryRow(ctx, query, employeeID, date, companyID).Scan(
		&att.ID, &att.EmployeeID, &att.CompanyID, &att.Date, &att.WorkScheduleTimeID, &att.ActualLocationType,
		&att.ClockIn, &att.ClockOut, &att.WorkHoursInMinutes,
		&att.ClockInLatitude, &att.ClockInLongitude, &att.ClockInProofURL,
		&att.ClockOutLatitude, &att.ClockOutLongitude, &att.ClockOutProofURL,
		&att.Status, &att.ApprovedBy, &att.ApprovedAt, &att.RejectionReason,
		&att.LeaveTypeID, &att.LateMinutes, &att.EarlyLeaveMinutes, &att.OvertimeMinutes,
		&att.CreatedAt, &att.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // No existing attendance found
		}
		return nil, fmt.Errorf("failed to get attendance by employee and date: %w", err)
	}

	return &att, nil
}

// GetByID implements attendance.AttendanceRepository.
func (a *attendanceRepository) GetByID(ctx context.Context, id string, companyID string) (attendance.Attendance, error) {
	q := GetQuerier(ctx, a.db)

	query := `
		SELECT 
			a.id, a.employee_id, a.company_id, a.date, a.work_schedule_time_id, a.actual_location_type,
			a.clock_in, a.clock_out, a.work_hours_in_minutes,
			a.clock_in_latitude, a.clock_in_longitude, a.clock_in_proof_url,
			a.clock_out_latitude, a.clock_out_longitude, a.clock_out_proof_url,
			a.status, a.approved_by, a.approved_at, a.rejection_reason,
			a.leave_type_id, a.late_minutes, a.early_leave_minutes, a.overtime_minutes,
			a.created_at, a.updated_at,
			e.full_name AS employee_name,
			p.name AS employee_position
		FROM attendances a
		LEFT JOIN employees e ON e.id = a.employee_id
		LEFT JOIN positions p ON p.id = e.position_id
		WHERE a.id = $1 AND a.company_id = $2
	`

	var att attendance.Attendance
	err := q.QueryRow(ctx, query, id, companyID).Scan(
		&att.ID, &att.EmployeeID, &att.CompanyID, &att.Date, &att.WorkScheduleTimeID, &att.ActualLocationType,
		&att.ClockIn, &att.ClockOut, &att.WorkHoursInMinutes,
		&att.ClockInLatitude, &att.ClockInLongitude, &att.ClockInProofURL,
		&att.ClockOutLatitude, &att.ClockOutLongitude, &att.ClockOutProofURL,
		&att.Status, &att.ApprovedBy, &att.ApprovedAt, &att.RejectionReason,
		&att.LeaveTypeID, &att.LateMinutes, &att.EarlyLeaveMinutes, &att.OvertimeMinutes,
		&att.CreatedAt, &att.UpdatedAt,
		&att.EmployeeName, &att.EmployeePosition,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return attendance.Attendance{}, attendance.ErrAttendanceNotFound
		}
		return attendance.Attendance{}, fmt.Errorf("failed to get attendance by ID: %w", err)
	}

	return att, nil
}

// GetMyAttendance implements attendance.AttendanceRepository.
func (a *attendanceRepository) GetMyAttendance(ctx context.Context, employeeID string, filter attendance.MyAttendanceFilter, companyID string) ([]attendance.Attendance, int64, error) {
	q := GetQuerier(ctx, a.db)

	// Build WHERE clause
	baseWhere := "a.employee_id = $1 AND a.company_id = $2"
	args := []interface{}{employeeID, companyID}
	argIdx := 3

	// Date filter
	if filter.Date != nil && *filter.Date != "" {
		baseWhere += fmt.Sprintf(" AND a.date = $%d", argIdx)
		args = append(args, *filter.Date)
		argIdx++
	}

	// Date range filters
	if filter.StartDate != nil && *filter.StartDate != "" {
		baseWhere += fmt.Sprintf(" AND a.date >= $%d", argIdx)
		args = append(args, *filter.StartDate)
		argIdx++
	}
	if filter.EndDate != nil && *filter.EndDate != "" {
		baseWhere += fmt.Sprintf(" AND a.date <= $%d", argIdx)
		args = append(args, *filter.EndDate)
		argIdx++
	}

	// Status filter
	if filter.Status != nil && *filter.Status != "" {
		baseWhere += fmt.Sprintf(" AND a.status = $%d", argIdx)
		args = append(args, *filter.Status)
		argIdx++
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM attendances a WHERE " + baseWhere
	var total int64
	if err := q.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count attendances: %w", err)
	}

	// Build ORDER BY
	orderByField := "a.date"
	switch filter.SortBy {
	case "clock_in_time":
		orderByField = "a.clock_in"
	case "clock_out_time":
		orderByField = "a.clock_out"
	case "status":
		orderByField = "a.status"
	}
	sortOrder := "DESC"
	if strings.ToLower(filter.SortOrder) == "asc" {
		sortOrder = "ASC"
	}

	// Build query with pagination
	selectQuery := fmt.Sprintf(`
		SELECT 
			a.id, a.employee_id, a.company_id, a.date, a.work_schedule_time_id, a.actual_location_type,
			a.clock_in, a.clock_out, a.work_hours_in_minutes,
			a.clock_in_latitude, a.clock_in_longitude, a.clock_in_proof_url,
			a.clock_out_latitude, a.clock_out_longitude, a.clock_out_proof_url,
			a.status, a.approved_by, a.approved_at, a.rejection_reason,
			a.leave_type_id, a.late_minutes, a.early_leave_minutes, a.overtime_minutes,
			a.created_at, a.updated_at,
			e.full_name AS employee_name,
			p.name AS employee_position
		FROM attendances a
		LEFT JOIN employees e ON e.id = a.employee_id
		LEFT JOIN positions p ON p.id = e.position_id
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, baseWhere, orderByField, sortOrder, argIdx, argIdx+1)

	limit := filter.Limit
	if limit == 0 {
		limit = 20
	}
	offset := (filter.Page - 1) * limit
	args = append(args, limit, offset)

	rows, err := q.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query attendances: %w", err)
	}
	defer rows.Close()

	var attendances []attendance.Attendance
	for rows.Next() {
		var att attendance.Attendance
		err := rows.Scan(
			&att.ID, &att.EmployeeID, &att.CompanyID, &att.Date, &att.WorkScheduleTimeID, &att.ActualLocationType,
			&att.ClockIn, &att.ClockOut, &att.WorkHoursInMinutes,
			&att.ClockInLatitude, &att.ClockInLongitude, &att.ClockInProofURL,
			&att.ClockOutLatitude, &att.ClockOutLongitude, &att.ClockOutProofURL,
			&att.Status, &att.ApprovedBy, &att.ApprovedAt, &att.RejectionReason,
			&att.LeaveTypeID, &att.LateMinutes, &att.EarlyLeaveMinutes, &att.OvertimeMinutes,
			&att.CreatedAt, &att.UpdatedAt,
			&att.EmployeeName, &att.EmployeePosition,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan attendance: %w", err)
		}
		attendances = append(attendances, att)
	}

	return attendances, total, nil
}

// HasCheckedInToday implements attendance.AttendanceRepository.
func (a *attendanceRepository) HasCheckedInToday(ctx context.Context, employeeID string, dateLocal string, companyID string) (bool, error) {
	q := GetQuerier(ctx, a.db)

	query := `
		SELECT EXISTS (
			SELECT 1
			FROM attendances
			WHERE employee_id = $1
			  AND date = $2
			  AND company_id = $3
		)
	`

	var hasCheckedIn bool
	err := q.QueryRow(ctx, query, employeeID, dateLocal, companyID).Scan(&hasCheckedIn)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, err
		}
		return false, fmt.Errorf("failed to check if employee has checked in today: %w", err)
	}

	return hasCheckedIn, nil
}

// List implements attendance.AttendanceRepository.
func (a *attendanceRepository) List(ctx context.Context, filter attendance.AttendanceFilter, companyID string) ([]attendance.Attendance, int64, error) {
	q := GetQuerier(ctx, a.db)

	// Build WHERE clause
	baseWhere := "a.company_id = $1"
	args := []interface{}{companyID}
	argIdx := 2

	// Employee ID filter
	if filter.EmployeeID != nil && *filter.EmployeeID != "" {
		baseWhere += fmt.Sprintf(" AND a.employee_id = $%d", argIdx)
		args = append(args, *filter.EmployeeID)
		argIdx++
	}

	// Employee name filter (search)
	if filter.EmployeeName != nil && *filter.EmployeeName != "" {
		baseWhere += fmt.Sprintf(" AND e.full_name ILIKE $%d", argIdx)
		args = append(args, "%"+*filter.EmployeeName+"%")
		argIdx++
	}

	// Date filter
	if filter.Date != nil && *filter.Date != "" {
		baseWhere += fmt.Sprintf(" AND a.date = $%d", argIdx)
		args = append(args, *filter.Date)
		argIdx++
	}

	// Date range filters
	if filter.StartDate != nil && *filter.StartDate != "" {
		baseWhere += fmt.Sprintf(" AND a.date >= $%d", argIdx)
		args = append(args, *filter.StartDate)
		argIdx++
	}
	if filter.EndDate != nil && *filter.EndDate != "" {
		baseWhere += fmt.Sprintf(" AND a.date <= $%d", argIdx)
		args = append(args, *filter.EndDate)
		argIdx++
	}

	// Status filter
	if filter.Status != nil && *filter.Status != "" {
		baseWhere += fmt.Sprintf(" AND a.status = $%d", argIdx)
		args = append(args, *filter.Status)
		argIdx++
	}

	// Count total (need to join employees for name filter)
	countQuery := `
		SELECT COUNT(*) 
		FROM attendances a
		LEFT JOIN employees e ON e.id = a.employee_id
		WHERE ` + baseWhere
	var total int64
	if err := q.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count attendances: %w", err)
	}

	// Build ORDER BY
	orderByField := "a.date"
	switch filter.SortBy {
	case "employee_name":
		orderByField = "e.full_name"
	case "clock_in_time":
		orderByField = "a.clock_in"
	case "clock_out_time":
		orderByField = "a.clock_out"
	case "status":
		orderByField = "a.status"
	}
	sortOrder := "DESC"
	if strings.ToLower(filter.SortOrder) == "asc" {
		sortOrder = "ASC"
	}

	// Build query with pagination
	selectQuery := fmt.Sprintf(`
		SELECT 
			a.id, a.employee_id, a.company_id, a.date, a.work_schedule_time_id, a.actual_location_type,
			a.clock_in, a.clock_out, a.work_hours_in_minutes,
			a.clock_in_latitude, a.clock_in_longitude, a.clock_in_proof_url,
			a.clock_out_latitude, a.clock_out_longitude, a.clock_out_proof_url,
			a.status, a.approved_by, a.approved_at, a.rejection_reason,
			a.leave_type_id, a.late_minutes, a.early_leave_minutes, a.overtime_minutes,
			a.created_at, a.updated_at,
			e.full_name AS employee_name,
			p.name AS employee_position
		FROM attendances a
		LEFT JOIN employees e ON e.id = a.employee_id
		LEFT JOIN positions p ON p.id = e.position_id
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, baseWhere, orderByField, sortOrder, argIdx, argIdx+1)

	limit := filter.Limit
	if limit == 0 {
		limit = 20
	}
	offset := (filter.Page - 1) * limit
	args = append(args, limit, offset)

	rows, err := q.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query attendances: %w", err)
	}
	defer rows.Close()

	var attendances []attendance.Attendance
	for rows.Next() {
		var att attendance.Attendance
		err := rows.Scan(
			&att.ID, &att.EmployeeID, &att.CompanyID, &att.Date, &att.WorkScheduleTimeID, &att.ActualLocationType,
			&att.ClockIn, &att.ClockOut, &att.WorkHoursInMinutes,
			&att.ClockInLatitude, &att.ClockInLongitude, &att.ClockInProofURL,
			&att.ClockOutLatitude, &att.ClockOutLongitude, &att.ClockOutProofURL,
			&att.Status, &att.ApprovedBy, &att.ApprovedAt, &att.RejectionReason,
			&att.LeaveTypeID, &att.LateMinutes, &att.EarlyLeaveMinutes, &att.OvertimeMinutes,
			&att.CreatedAt, &att.UpdatedAt,
			&att.EmployeeName, &att.EmployeePosition,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan attendance: %w", err)
		}
		attendances = append(attendances, att)
	}

	return attendances, total, nil
}

// Update implements attendance.AttendanceRepository.
func (a *attendanceRepository) Update(ctx context.Context, att attendance.Attendance) error {
	q := GetQuerier(ctx, a.db)

	updates := make([]string, 0)
	args := make([]interface{}, 0)
	argIdx := 1

	if att.ClockIn != nil {
		updates = append(updates, fmt.Sprintf("clock_in = $%d", argIdx))
		args = append(args, att.ClockIn)
		argIdx++
	}
	if att.ClockOut != nil {
		updates = append(updates, fmt.Sprintf("clock_out = $%d", argIdx))
		args = append(args, att.ClockOut)
		argIdx++
	}
	if att.WorkHoursInMinutes != nil {
		updates = append(updates, fmt.Sprintf("work_hours_in_minutes = $%d", argIdx))
		args = append(args, att.WorkHoursInMinutes)
		argIdx++
	}
	if att.ClockInLatitude != nil {
		updates = append(updates, fmt.Sprintf("clock_in_latitude = $%d", argIdx))
		args = append(args, att.ClockInLatitude)
		argIdx++
	}
	if att.ClockInLongitude != nil {
		updates = append(updates, fmt.Sprintf("clock_in_longitude = $%d", argIdx))
		args = append(args, att.ClockInLongitude)
		argIdx++
	}
	if att.ClockInProofURL != nil {
		updates = append(updates, fmt.Sprintf("clock_in_proof_url = $%d", argIdx))
		args = append(args, att.ClockInProofURL)
		argIdx++
	}
	if att.ClockOutLatitude != nil {
		updates = append(updates, fmt.Sprintf("clock_out_latitude = $%d", argIdx))
		args = append(args, att.ClockOutLatitude)
		argIdx++
	}
	if att.ClockOutLongitude != nil {
		updates = append(updates, fmt.Sprintf("clock_out_longitude = $%d", argIdx))
		args = append(args, att.ClockOutLongitude)
		argIdx++
	}
	if att.ClockOutProofURL != nil {
		updates = append(updates, fmt.Sprintf("clock_out_proof_url = $%d", argIdx))
		args = append(args, att.ClockOutProofURL)
		argIdx++
	}
	if att.Status != "" {
		updates = append(updates, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, att.Status)
		argIdx++
	}
	if att.ApprovedBy != nil {
		updates = append(updates, fmt.Sprintf("approved_by = $%d", argIdx))
		args = append(args, att.ApprovedBy)
		argIdx++
	}
	if att.ApprovedAt != nil {
		updates = append(updates, fmt.Sprintf("approved_at = $%d", argIdx))
		args = append(args, att.ApprovedAt)
		argIdx++
	}
	if att.RejectionReason != nil {
		updates = append(updates, fmt.Sprintf("rejection_reason = $%d", argIdx))
		args = append(args, att.RejectionReason)
		argIdx++
	}
	if att.LateMinutes != nil {
		updates = append(updates, fmt.Sprintf("late_minutes = $%d", argIdx))
		args = append(args, att.LateMinutes)
		argIdx++
	}
	if att.EarlyLeaveMinutes != nil {
		updates = append(updates, fmt.Sprintf("early_leave_minutes = $%d", argIdx))
		args = append(args, att.EarlyLeaveMinutes)
		argIdx++
	}
	if att.OvertimeMinutes != nil {
		updates = append(updates, fmt.Sprintf("overtime_minutes = $%d", argIdx))
		args = append(args, att.OvertimeMinutes)
		argIdx++
	}

	if len(updates) == 0 {
		return fmt.Errorf("no updatable fields provided for attendance update")
	}

	updates = append(updates, fmt.Sprintf("updated_at = $%d", argIdx))
	args = append(args, time.Now())
	argIdx++

	args = append(args, att.ID)
	idIdx := argIdx
	argIdx++

	args = append(args, att.CompanyID)

	query := "UPDATE attendances SET " + strings.Join(updates, ", ") +
		fmt.Sprintf(" WHERE id = $%d AND company_id = $%d RETURNING id", idIdx, argIdx)

	var updatedID string
	if err := q.QueryRow(ctx, query, args...).Scan(&updatedID); err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("attendance not found: %w", err)
		}
		return fmt.Errorf("failed to update attendance: %w", err)
	}

	return nil
}

// Delete implements attendance.AttendanceRepository.
func (a *attendanceRepository) Delete(ctx context.Context, id string, companyID string) error {
	q := GetQuerier(ctx, a.db)

	query := `DELETE FROM attendances WHERE id = $1 AND company_id = $2`

	commandTag, err := q.Exec(ctx, query, id, companyID)
	if err != nil {
		return fmt.Errorf("failed to delete attendance: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return attendance.ErrAttendanceNotFound
	}

	return nil
}

func NewAttendanceRepository(db *database.DB) attendance.AttendanceRepository {
	return &attendanceRepository{db: db}
}
