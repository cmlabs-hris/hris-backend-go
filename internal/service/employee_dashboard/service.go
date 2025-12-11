package employee_dashboard

import (
	"context"
	"fmt"
	"strconv"
	"time"

	empDashboard "github.com/cmlabs-hris/hris-backend-go/internal/domain/employee_dashboard"
	"github.com/go-chi/jwtauth/v5"
	"golang.org/x/sync/errgroup"
)

type EmployeeDashboardServiceImpl struct {
	empDashboard.EmployeeDashboardRepository
}

func NewEmployeeDashboardService(repo empDashboard.EmployeeDashboardRepository) empDashboard.EmployeeDashboardService {
	return &EmployeeDashboardServiceImpl{
		EmployeeDashboardRepository: repo,
	}
}

// getEmployeeID extracts employee_id from JWT claims
func (s *EmployeeDashboardServiceImpl) getEmployeeID(ctx context.Context) (string, error) {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to extract claims from context: %w", err)
	}

	employeeID, ok := claims["employee_id"].(string)
	if !ok || employeeID == "" {
		return "", fmt.Errorf("employee_id not found in claims")
	}
	return employeeID, nil
}

// formatWorkHours formats minutes to "Xh Ym" format
func formatWorkHours(minutes int64) string {
	hours := minutes / 60
	mins := minutes % 60
	return fmt.Sprintf("%dh %dm", hours, mins)
}

// parseMonth parses YYYY-MM format, defaults to current month
func parseMonth(month string) (int, int) {
	now := time.Now()
	if month == "" {
		return now.Year(), int(now.Month())
	}

	parsed, err := time.Parse("2006-01", month)
	if err != nil {
		return now.Year(), int(now.Month())
	}
	return parsed.Year(), int(parsed.Month())
}

// parseYear parses year string, defaults to current year
func parseYear(year string) int {
	now := time.Now()
	if year == "" {
		return now.Year()
	}

	y, err := strconv.Atoi(year)
	if err != nil || y < 2000 || y > 2100 {
		return now.Year()
	}
	return y
}

// parseWeek parses week number, defaults to current week of month
func parseWeek(week string) int {
	if week == "" {
		now := time.Now()
		// Calculate week of month (1-based)
		day := now.Day()
		return ((day - 1) / 7) + 1
	}

	w, err := strconv.Atoi(week)
	if err != nil || w < 1 || w > 5 {
		now := time.Now()
		day := now.Day()
		return ((day - 1) / 7) + 1
	}
	return w
}

// parseDateRange parses start and end date, defaults to current month
func parseDateRange(startDate, endDate string) (time.Time, time.Time) {
	now := time.Now()

	// Default: current month
	defaultStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	defaultEnd := defaultStart.AddDate(0, 1, 0)

	if startDate == "" || endDate == "" {
		return defaultStart, defaultEnd
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return defaultStart, defaultEnd
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return defaultStart, defaultEnd
	}

	// End date should be exclusive (add 1 day)
	end = end.AddDate(0, 0, 1)

	return start, end
}

// GetDashboard returns combined employee dashboard data
func (s *EmployeeDashboardServiceImpl) GetDashboard(ctx context.Context) (*empDashboard.EmployeeDashboardResponse, error) {
	employeeID, err := s.getEmployeeID(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	year, month := now.Year(), int(now.Month())
	week := ((now.Day() - 1) / 7) + 1

	// Default date range: current month
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	var (
		workStats         empDashboard.WorkStatsResponse
		attendanceSummary empDashboard.AttendanceSummaryResponse
		leaveSummary      empDashboard.LeaveSummaryResponse
		workHoursChart    empDashboard.WorkHoursChartResponse
	)

	g, gCtx := errgroup.WithContext(ctx)

	// 1. Work Stats
	g.Go(func() error {
		data, err := s.EmployeeDashboardRepository.GetWorkStats(gCtx, employeeID, startDate, endDate)
		if err != nil {
			return err
		}
		workStats = empDashboard.WorkStatsResponse{
			WorkHours:   formatWorkHours(data.TotalWorkMinutes),
			WorkMinutes: data.TotalWorkMinutes,
			OnTimeCount: data.OnTimeCount,
			LateCount:   data.LateCount,
			AbsentCount: data.AbsentCount,
			StartDate:   startDate.Format("2006-01-02"),
			EndDate:     endDate.AddDate(0, 0, -1).Format("2006-01-02"),
		}
		return nil
	})

	// 2. Attendance Summary
	g.Go(func() error {
		data, err := s.EmployeeDashboardRepository.GetAttendanceSummary(gCtx, employeeID, year, month)
		if err != nil {
			return err
		}
		attendanceSummary = s.buildAttendanceSummaryResponse(data, year, month)
		return nil
	})

	// 3. Leave Summary
	g.Go(func() error {
		data, err := s.EmployeeDashboardRepository.GetLeaveSummary(gCtx, employeeID, year)
		if err != nil {
			return err
		}
		leaveSummary = s.buildLeaveSummaryResponse(data, year)
		return nil
	})

	// 4. Work Hours Chart
	g.Go(func() error {
		data, err := s.EmployeeDashboardRepository.GetWorkHoursChart(gCtx, employeeID, year, month, week)
		if err != nil {
			return err
		}
		workHoursChart = s.buildWorkHoursChartResponse(data, year, month, week)
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return &empDashboard.EmployeeDashboardResponse{
		WorkStats:         workStats,
		AttendanceSummary: attendanceSummary,
		LeaveSummary:      leaveSummary,
		WorkHoursChart:    workHoursChart,
	}, nil
}

// GetWorkStats returns work stats for a date range
func (s *EmployeeDashboardServiceImpl) GetWorkStats(ctx context.Context, startDateStr, endDateStr string) (*empDashboard.WorkStatsResponse, error) {
	employeeID, err := s.getEmployeeID(ctx)
	if err != nil {
		return nil, err
	}

	startDate, endDate := parseDateRange(startDateStr, endDateStr)

	data, err := s.EmployeeDashboardRepository.GetWorkStats(ctx, employeeID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	return &empDashboard.WorkStatsResponse{
		WorkHours:   formatWorkHours(data.TotalWorkMinutes),
		WorkMinutes: data.TotalWorkMinutes,
		OnTimeCount: data.OnTimeCount,
		LateCount:   data.LateCount,
		AbsentCount: data.AbsentCount,
		StartDate:   startDate.Format("2006-01-02"),
		EndDate:     endDate.AddDate(0, 0, -1).Format("2006-01-02"),
	}, nil
}

// GetAttendanceSummary returns attendance summary for a month
func (s *EmployeeDashboardServiceImpl) GetAttendanceSummary(ctx context.Context, month string) (*empDashboard.AttendanceSummaryResponse, error) {
	employeeID, err := s.getEmployeeID(ctx)
	if err != nil {
		return nil, err
	}

	year, m := parseMonth(month)

	data, err := s.EmployeeDashboardRepository.GetAttendanceSummary(ctx, employeeID, year, m)
	if err != nil {
		return nil, err
	}

	result := s.buildAttendanceSummaryResponse(data, year, m)
	return &result, nil
}

// GetLeaveSummary returns leave quota summary for a year
func (s *EmployeeDashboardServiceImpl) GetLeaveSummary(ctx context.Context, yearStr string) (*empDashboard.LeaveSummaryResponse, error) {
	employeeID, err := s.getEmployeeID(ctx)
	if err != nil {
		return nil, err
	}

	year := parseYear(yearStr)

	data, err := s.EmployeeDashboardRepository.GetLeaveSummary(ctx, employeeID, year)
	if err != nil {
		return nil, err
	}

	result := s.buildLeaveSummaryResponse(data, year)
	return &result, nil
}

// GetWorkHoursChart returns daily work hours for a specific week
func (s *EmployeeDashboardServiceImpl) GetWorkHoursChart(ctx context.Context, weekStr string) (*empDashboard.WorkHoursChartResponse, error) {
	employeeID, err := s.getEmployeeID(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	year, month := now.Year(), int(now.Month())
	week := parseWeek(weekStr)

	data, err := s.EmployeeDashboardRepository.GetWorkHoursChart(ctx, employeeID, year, month, week)
	if err != nil {
		return nil, err
	}

	result := s.buildWorkHoursChartResponse(data, year, month, week)
	return &result, nil
}

// buildAttendanceSummaryResponse builds response with percentages
func (s *EmployeeDashboardServiceImpl) buildAttendanceSummaryResponse(data *empDashboard.AttendanceSummaryData, year, month int) empDashboard.AttendanceSummaryResponse {
	total := data.OnTime + data.Late + data.Absent + data.LeaveCount

	var onTimePercent, latePercent, absentPercent, leavePercent float64
	if total > 0 {
		onTimePercent = float64(data.OnTime) / float64(total) * 100
		latePercent = float64(data.Late) / float64(total) * 100
		absentPercent = float64(data.Absent) / float64(total) * 100
		leavePercent = float64(data.LeaveCount) / float64(total) * 100
	}

	var leaveBreakdown []empDashboard.LeaveBreakdownItem
	for _, item := range data.LeaveBreakdown {
		var percent float64
		if total > 0 {
			percent = float64(item.Count) / float64(total) * 100
		}
		leaveBreakdown = append(leaveBreakdown, empDashboard.LeaveBreakdownItem{
			LeaveTypeName: item.LeaveTypeName,
			Count:         item.Count,
			Percent:       percent,
		})
	}

	return empDashboard.AttendanceSummaryResponse{
		TotalAttendance: total,
		OnTime:          data.OnTime,
		Late:            data.Late,
		Absent:          data.Absent,
		LeaveCount:      data.LeaveCount,
		OnTimePercent:   onTimePercent,
		LatePercent:     latePercent,
		AbsentPercent:   absentPercent,
		LeavePercent:    leavePercent,
		LeaveBreakdown:  leaveBreakdown,
		Month:           fmt.Sprintf("%d-%02d", year, month),
	}
}

// buildLeaveSummaryResponse builds response from quota data
func (s *EmployeeDashboardServiceImpl) buildLeaveSummaryResponse(data []empDashboard.LeaveQuotaData, year int) empDashboard.LeaveSummaryResponse {
	var items []empDashboard.LeaveQuotaItem
	for _, d := range data {
		items = append(items, empDashboard.LeaveQuotaItem{
			LeaveTypeID:   d.LeaveTypeID,
			LeaveTypeName: d.LeaveTypeName,
			TotalQuota:    d.TotalQuota,
			Taken:         d.UsedQuota,
			Remaining:     d.AvailableQuota,
		})
	}

	return empDashboard.LeaveSummaryResponse{
		Year:             year,
		LeaveQuotaDetail: items,
	}
}

// buildWorkHoursChartResponse builds response with daily breakdown
func (s *EmployeeDashboardServiceImpl) buildWorkHoursChartResponse(data []empDashboard.DailyWorkHourData, year, month, week int) empDashboard.WorkHoursChartResponse {
	var totalMinutes int64
	var dailyItems []empDashboard.DailyWorkHourItem

	for _, d := range data {
		totalMinutes += d.WorkMinutes
		dailyItems = append(dailyItems, empDashboard.DailyWorkHourItem{
			Date:        d.Date.Format("2006-01-02"),
			DayName:     d.Date.Weekday().String(),
			WorkHours:   formatWorkHours(d.WorkMinutes),
			WorkMinutes: d.WorkMinutes,
		})
	}

	return empDashboard.WorkHoursChartResponse{
		TotalWorkHours:   formatWorkHours(totalMinutes),
		TotalWorkMinutes: totalMinutes,
		WeekNumber:       week,
		Year:             year,
		Month:            month,
		DailyWorkHours:   dailyItems,
	}
}
