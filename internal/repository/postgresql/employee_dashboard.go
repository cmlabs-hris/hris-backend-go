package postgresql

import (
	"context"
	"fmt"
	"time"

	empDashboard "github.com/cmlabs-hris/hris-backend-go/internal/domain/employee_dashboard"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
)

type employeeDashboardRepositoryImpl struct {
	db *database.DB
}

func NewEmployeeDashboardRepository(db *database.DB) empDashboard.EmployeeDashboardRepository {
	return &employeeDashboardRepositoryImpl{db: db}
}

// GetWorkStats returns work hours and attendance counts for a date range (single query)
func (r *employeeDashboardRepositoryImpl) GetWorkStats(ctx context.Context, employeeID string, startDate, endDate time.Time) (*empDashboard.WorkStatsData, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT 
			COALESCE(SUM(work_hours_in_minutes), 0) as total_work_minutes,
			COALESCE(SUM(CASE WHEN status = 'on_time' THEN 1 ELSE 0 END), 0) as on_time_count,
			COALESCE(SUM(CASE WHEN status = 'late' THEN 1 ELSE 0 END), 0) as late_count,
			COALESCE(SUM(CASE WHEN status = 'absent' THEN 1 ELSE 0 END), 0) as absent_count
		FROM attendances
		WHERE employee_id = $1
		AND date >= $2 AND date < $3
	`

	var data empDashboard.WorkStatsData
	err := q.QueryRow(ctx, query, employeeID, startDate, endDate).Scan(
		&data.TotalWorkMinutes, &data.OnTimeCount, &data.LateCount, &data.AbsentCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get work stats: %w", err)
	}
	return &data, nil
}

// GetAttendanceSummary returns attendance distribution for a month (2 queries optimized)
func (r *employeeDashboardRepositoryImpl) GetAttendanceSummary(ctx context.Context, employeeID string, year, month int) (*empDashboard.AttendanceSummaryData, error) {
	q := GetQuerier(ctx, r.db)

	startOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	// Query 1: Get attendance counts + leave count with breakdown
	countQuery := `
		SELECT 
			COALESCE(SUM(CASE WHEN status = 'on_time' THEN 1 ELSE 0 END), 0) as on_time,
			COALESCE(SUM(CASE WHEN status = 'late' THEN 1 ELSE 0 END), 0) as late,
			COALESCE(SUM(CASE WHEN status = 'absent' THEN 1 ELSE 0 END), 0) as absent,
			COALESCE(SUM(CASE WHEN status = 'leave' THEN 1 ELSE 0 END), 0) as leave_count
		FROM attendances
		WHERE employee_id = $1
		AND date >= $2 AND date < $3
	`

	var data empDashboard.AttendanceSummaryData
	err := q.QueryRow(ctx, countQuery, employeeID, startOfMonth, endOfMonth).Scan(
		&data.OnTime, &data.Late, &data.Absent, &data.LeaveCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get attendance summary counts: %w", err)
	}

	// Query 2: Get leave breakdown by type
	breakdownQuery := `
		SELECT 
			COALESCE(a.leave_type_id::text, '') as leave_type_id,
			COALESCE(lt.name, 'Unknown') as leave_type_name,
			COUNT(*) as count
		FROM attendances a
		LEFT JOIN leave_types lt ON a.leave_type_id = lt.id
		WHERE a.employee_id = $1
		AND a.date >= $2 AND a.date < $3
		AND a.status = 'leave'
		AND a.leave_type_id IS NOT NULL
		GROUP BY a.leave_type_id, lt.name
		ORDER BY count DESC
	`

	rows, err := q.Query(ctx, breakdownQuery, employeeID, startOfMonth, endOfMonth)
	if err != nil {
		return nil, fmt.Errorf("failed to get leave breakdown: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item empDashboard.LeaveBreakdownData
		if err := rows.Scan(&item.LeaveTypeID, &item.LeaveTypeName, &item.Count); err != nil {
			return nil, fmt.Errorf("failed to scan leave breakdown: %w", err)
		}
		data.LeaveBreakdown = append(data.LeaveBreakdown, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &data, nil
}

// GetLeaveSummary returns leave quota summary for a year (single query)
func (r *employeeDashboardRepositoryImpl) GetLeaveSummary(ctx context.Context, employeeID string, year int) ([]empDashboard.LeaveQuotaData, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT 
			lq.leave_type_id,
			lt.name as leave_type_name,
			COALESCE(lq.earned_quota, 0) + COALESCE(lq.rollover_quota, 0) + COALESCE(lq.adjustment_quota, 0) as total_quota,
			COALESCE(lq.used_quota, 0) as used_quota,
			COALESCE(lq.available_quota, 0) as available_quota
		FROM leave_quotas lq
		JOIN leave_types lt ON lq.leave_type_id = lt.id
		WHERE lq.employee_id = $1
		AND lq.year = $2
		ORDER BY lt.name
	`

	rows, err := q.Query(ctx, query, employeeID, year)
	if err != nil {
		return nil, fmt.Errorf("failed to get leave summary: %w", err)
	}
	defer rows.Close()

	var result []empDashboard.LeaveQuotaData
	for rows.Next() {
		var item empDashboard.LeaveQuotaData
		if err := rows.Scan(&item.LeaveTypeID, &item.LeaveTypeName, &item.TotalQuota, &item.UsedQuota, &item.AvailableQuota); err != nil {
			return nil, fmt.Errorf("failed to scan leave quota: %w", err)
		}
		result = append(result, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// GetWorkHoursChart returns daily work hours for a specific week (single query)
func (r *employeeDashboardRepositoryImpl) GetWorkHoursChart(ctx context.Context, employeeID string, year, month, week int) ([]empDashboard.DailyWorkHourData, error) {
	q := GetQuerier(ctx, r.db)

	// Calculate week date range
	startOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	weekStartDay := (week - 1) * 7
	startOfWeek := startOfMonth.AddDate(0, 0, weekStartDay)
	endOfWeek := startOfWeek.AddDate(0, 0, 7)

	// Clamp to month boundaries
	endOfMonth := startOfMonth.AddDate(0, 1, 0)
	if endOfWeek.After(endOfMonth) {
		endOfWeek = endOfMonth
	}

	query := `
		SELECT 
			date,
			COALESCE(work_hours_in_minutes, 0) as work_minutes
		FROM attendances
		WHERE employee_id = $1
		AND date >= $2 AND date < $3
		ORDER BY date ASC
	`

	rows, err := q.Query(ctx, query, employeeID, startOfWeek, endOfWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get work hours chart: %w", err)
	}
	defer rows.Close()

	var result []empDashboard.DailyWorkHourData
	for rows.Next() {
		var item empDashboard.DailyWorkHourData
		if err := rows.Scan(&item.Date, &item.WorkMinutes); err != nil {
			return nil, fmt.Errorf("failed to scan daily work hours: %w", err)
		}
		result = append(result, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
