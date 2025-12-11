package postgresql

import (
	"context"
	"fmt"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/dashboard"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
)

type dashboardRepositoryImpl struct {
	db *database.DB
}

func NewDashboardRepository(db *database.DB) dashboard.DashboardRepository {
	return &dashboardRepositoryImpl{db: db}
}

// GetEmployeeSummary returns total, new (since date), active, resigned in single query
func (r *dashboardRepositoryImpl) GetEmployeeSummary(ctx context.Context, companyID string, since time.Time) (*dashboard.EmployeeSummaryStats, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT 
			COUNT(*) as total,
			COALESCE(SUM(CASE WHEN hire_date >= $2 THEN 1 ELSE 0 END), 0) as new_count,
			COALESCE(SUM(CASE WHEN employment_status = 'active' THEN 1 ELSE 0 END), 0) as active_count,
			COALESCE(SUM(CASE WHEN employment_status = 'resigned' THEN 1 ELSE 0 END), 0) as resigned_count
		FROM employees 
		WHERE company_id = $1 AND deleted_at IS NULL
	`

	var stats dashboard.EmployeeSummaryStats
	err := q.QueryRow(ctx, query, companyID, since).Scan(
		&stats.Total, &stats.New, &stats.Active, &stats.Resigned,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee summary: %w", err)
	}
	return &stats, nil
}

// GetEmployeeMonthlyStats returns new/active/resign for a month in single query
func (r *dashboardRepositoryImpl) GetEmployeeMonthlyStats(ctx context.Context, companyID string, year, month int) (*dashboard.EmployeeMonthlyStats, error) {
	q := GetQuerier(ctx, r.db)

	startOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN hire_date >= $2 AND hire_date < $3 THEN 1 ELSE 0 END), 0) as new_count,
			COALESCE(SUM(CASE WHEN employment_status = 'active' AND hire_date <= $3 AND (resignation_date IS NULL OR resignation_date > $3) THEN 1 ELSE 0 END), 0) as active_count,
			COALESCE(SUM(CASE WHEN employment_status = 'resigned' AND resignation_date >= $2 AND resignation_date < $3 THEN 1 ELSE 0 END), 0) as resign_count
		FROM employees 
		WHERE company_id = $1 AND deleted_at IS NULL
	`

	var stats dashboard.EmployeeMonthlyStats
	err := q.QueryRow(ctx, query, companyID, startOfMonth, endOfMonth).Scan(
		&stats.New, &stats.Active, &stats.Resign,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee monthly stats: %w", err)
	}
	return &stats, nil
}

// GetEmployeeTypeStats returns employment type distribution in single query
func (r *dashboardRepositoryImpl) GetEmployeeTypeStats(ctx context.Context, companyID string, year, month int) (*dashboard.EmployeeTypeStats, error) {
	q := GetQuerier(ctx, r.db)

	endOfMonth := time.Date(year, time.Month(month)+1, 0, 23, 59, 59, 0, time.UTC)

	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN employment_type = 'permanent' THEN 1 ELSE 0 END), 0) as permanent,
			COALESCE(SUM(CASE WHEN employment_type = 'probation' THEN 1 ELSE 0 END), 0) as probation,
			COALESCE(SUM(CASE WHEN employment_type = 'contract' THEN 1 ELSE 0 END), 0) as contract,
			COALESCE(SUM(CASE WHEN employment_type = 'internship' THEN 1 ELSE 0 END), 0) as internship,
			COALESCE(SUM(CASE WHEN employment_type = 'freelance' THEN 1 ELSE 0 END), 0) as freelance
		FROM employees 
		WHERE company_id = $1 AND deleted_at IS NULL 
		AND employment_status = 'active'
		AND hire_date <= $2
		AND (resignation_date IS NULL OR resignation_date > $2)
	`

	var stats dashboard.EmployeeTypeStats
	err := q.QueryRow(ctx, query, companyID, endOfMonth).Scan(
		&stats.Permanent, &stats.Probation, &stats.Contract, &stats.Internship, &stats.Freelance,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee type stats: %w", err)
	}
	return &stats, nil
}

// GetAttendanceStatsByDay returns on_time/late/absent for a specific day
func (r *dashboardRepositoryImpl) GetAttendanceStatsByDay(ctx context.Context, companyID string, date time.Time) (*dashboard.AttendanceStats, error) {
	q := GetQuerier(ctx, r.db)

	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.AddDate(0, 0, 1)

	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN status = 'on_time' THEN 1 ELSE 0 END), 0) as on_time,
			COALESCE(SUM(CASE WHEN status = 'late' THEN 1 ELSE 0 END), 0) as late,
			COALESCE(SUM(CASE WHEN status = 'absent' THEN 1 ELSE 0 END), 0) as absent
		FROM attendances 
		WHERE company_id = $1 
		AND date >= $2 AND date < $3
	`

	var stats dashboard.AttendanceStats
	err := q.QueryRow(ctx, query, companyID, startOfDay, endOfDay).Scan(
		&stats.OnTime, &stats.Late, &stats.Absent,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get attendance stats by day: %w", err)
	}
	return &stats, nil
}

// GetMonthlyAttendanceWithRecords returns monthly stats + latest records
// Uses 2 queries but they run in the same DB call context
func (r *dashboardRepositoryImpl) GetMonthlyAttendanceWithRecords(ctx context.Context, companyID string, year, month int, limit int) (*dashboard.MonthlyAttendanceData, error) {
	q := GetQuerier(ctx, r.db)

	startOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	// Query 1: Get counts
	countQuery := `
		SELECT 
			COALESCE(SUM(CASE WHEN status = 'on_time' THEN 1 ELSE 0 END), 0) as on_time,
			COALESCE(SUM(CASE WHEN status = 'late' THEN 1 ELSE 0 END), 0) as late,
			COALESCE(SUM(CASE WHEN status = 'absent' THEN 1 ELSE 0 END), 0) as absent
		FROM attendances 
		WHERE company_id = $1 
		AND date >= $2 AND date < $3
	`

	var data dashboard.MonthlyAttendanceData
	err := q.QueryRow(ctx, countQuery, companyID, startOfMonth, endOfMonth).Scan(
		&data.OnTime, &data.Late, &data.Absent,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly attendance stats: %w", err)
	}

	// Query 2: Get latest records
	recordsQuery := `
		SELECT e.full_name, a.status, a.clock_in
		FROM attendances a
		JOIN employees e ON a.employee_id = e.id
		WHERE a.company_id = $1 
		AND a.date >= $2 AND a.date < $3
		ORDER BY a.created_at DESC
		LIMIT $4
	`

	rows, err := q.Query(ctx, recordsQuery, companyID, startOfMonth, endOfMonth, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest attendance records: %w", err)
	}
	defer rows.Close()

	no := 1
	for rows.Next() {
		var record dashboard.AttendanceRecordItem
		var clockIn *time.Time
		if err := rows.Scan(&record.EmployeeName, &record.Status, &clockIn); err != nil {
			return nil, fmt.Errorf("failed to scan attendance record: %w", err)
		}
		record.No = no
		if clockIn != nil {
			checkInStr := clockIn.Format("15:04")
			record.CheckIn = &checkInStr
		}
		data.Records = append(data.Records, record)
		no++
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &data, nil
}
