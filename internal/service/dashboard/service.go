package dashboard

import (
	"context"
	"fmt"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/dashboard"
	"github.com/go-chi/jwtauth/v5"
	"golang.org/x/sync/errgroup"
)

type DashboardServiceImpl struct {
	dashboard.DashboardRepository
}

func NewDashboardService(repo dashboard.DashboardRepository) dashboard.DashboardService {
	return &DashboardServiceImpl{
		DashboardRepository: repo,
	}
}

// getCompanyID extracts company_id from JWT claims
func (s *DashboardServiceImpl) getCompanyID(ctx context.Context) (string, error) {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to extract claims from context: %w", err)
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		return "", fmt.Errorf("company_id not found in claims")
	}
	return companyID, nil
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

// parseDate parses YYYY-MM-DD format, defaults to today
func parseDate(date string) time.Time {
	now := time.Now()
	if date == "" {
		return now
	}

	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		return now
	}
	return parsed
}

// GetDashboard returns combined dashboard data using parallel goroutines
// Optimized: 5 goroutines, each with 1 DB query = 5 total queries (was 11+)
func (s *DashboardServiceImpl) GetDashboard(ctx context.Context) (*dashboard.DashboardResponse, error) {
	companyID, err := s.getCompanyID(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	year, month := now.Year(), int(now.Month())
	since := now.AddDate(0, 0, -30)

	var (
		employeeSummary   dashboard.EmployeeSummaryResponse
		currentNumber     dashboard.EmployeeCurrentNumberResponse
		statusStats       dashboard.EmployeeStatusStatsResponse
		attendanceStats   dashboard.AttendanceStatsResponse
		monthlyAttendance dashboard.MonthlyAttendanceResponse
	)

	g, gCtx := errgroup.WithContext(ctx)

	// 1. Employee Summary (1 query: total, new, active, resigned)
	g.Go(func() error {
		stats, err := s.GetEmployeeSummary(gCtx, companyID, since)
		if err != nil {
			return err
		}
		employeeSummary = dashboard.EmployeeSummaryResponse{
			TotalEmployee:    stats.Total,
			NewEmployee:      stats.New,
			ActiveEmployee:   stats.Active,
			ResignedEmployee: stats.Resigned,
		}
		return nil
	})

	// 2. Employee Current Number (1 query: new, active, resign for month)
	g.Go(func() error {
		stats, err := s.GetEmployeeMonthlyStats(gCtx, companyID, year, month)
		if err != nil {
			return err
		}
		currentNumber = dashboard.EmployeeCurrentNumberResponse{
			New:    stats.New,
			Active: stats.Active,
			Resign: stats.Resign,
		}
		return nil
	})

	// 3. Employee Status Stats (1 query: all employment types)
	g.Go(func() error {
		stats, err := s.GetEmployeeTypeStats(gCtx, companyID, year, month)
		if err != nil {
			return err
		}
		statusStats = dashboard.EmployeeStatusStatsResponse{
			Permanent:  stats.Permanent,
			Probation:  stats.Probation,
			Contract:   stats.Contract,
			Internship: stats.Internship,
			Freelance:  stats.Freelance,
		}
		return nil
	})

	// 4. Daily Attendance Stats (1 query)
	g.Go(func() error {
		stats, err := s.GetAttendanceStatsByDay(gCtx, companyID, now)
		if err != nil {
			return err
		}
		total := stats.OnTime + stats.Late + stats.Absent
		var onTimePercent, latePercent, absentPercent float64
		if total > 0 {
			onTimePercent = float64(stats.OnTime) / float64(total) * 100
			latePercent = float64(stats.Late) / float64(total) * 100
			absentPercent = float64(stats.Absent) / float64(total) * 100
		}
		attendanceStats = dashboard.AttendanceStatsResponse{
			OnTime:        stats.OnTime,
			Late:          stats.Late,
			Absent:        stats.Absent,
			OnTimePercent: onTimePercent,
			LatePercent:   latePercent,
			AbsentPercent: absentPercent,
		}
		return nil
	})

	// 5. Monthly Attendance with Records (2 queries in same call)
	g.Go(func() error {
		data, err := s.GetMonthlyAttendanceWithRecords(gCtx, companyID, year, month, 10)
		if err != nil {
			return err
		}
		monthlyAttendance = dashboard.MonthlyAttendanceResponse{
			OnTime:  data.OnTime,
			Late:    data.Late,
			Absent:  data.Absent,
			Records: data.Records,
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return &dashboard.DashboardResponse{
		EmployeeSummary:       employeeSummary,
		EmployeeCurrentNumber: currentNumber,
		EmployeeStatusStats:   statusStats,
		AttendanceStats:       attendanceStats,
		MonthlyAttendance:     monthlyAttendance,
	}, nil
}

// GetEmployeeCurrentNumber returns new/active/resign counts for a month (1 query)
func (s *DashboardServiceImpl) GetEmployeeCurrentNumber(ctx context.Context, month string) (*dashboard.EmployeeCurrentNumberResponse, error) {
	companyID, err := s.getCompanyID(ctx)
	if err != nil {
		return nil, err
	}

	year, m := parseMonth(month)

	stats, err := s.GetEmployeeMonthlyStats(ctx, companyID, year, m)
	if err != nil {
		return nil, err
	}

	return &dashboard.EmployeeCurrentNumberResponse{
		New:    stats.New,
		Active: stats.Active,
		Resign: stats.Resign,
	}, nil
}

// GetEmployeeStatusStats returns employee counts by employment type (1 query)
func (s *DashboardServiceImpl) GetEmployeeStatusStats(ctx context.Context, month string) (*dashboard.EmployeeStatusStatsResponse, error) {
	companyID, err := s.getCompanyID(ctx)
	if err != nil {
		return nil, err
	}

	year, m := parseMonth(month)

	stats, err := s.GetEmployeeTypeStats(ctx, companyID, year, m)
	if err != nil {
		return nil, err
	}

	return &dashboard.EmployeeStatusStatsResponse{
		Permanent:  stats.Permanent,
		Probation:  stats.Probation,
		Contract:   stats.Contract,
		Internship: stats.Internship,
		Freelance:  stats.Freelance,
	}, nil
}

// GetMonthlyAttendance returns attendance stats and latest records for a month (2 queries)
func (s *DashboardServiceImpl) GetMonthlyAttendance(ctx context.Context, month string) (*dashboard.MonthlyAttendanceResponse, error) {
	companyID, err := s.getCompanyID(ctx)
	if err != nil {
		return nil, err
	}

	year, m := parseMonth(month)

	data, err := s.GetMonthlyAttendanceWithRecords(ctx, companyID, year, m, 10)
	if err != nil {
		return nil, err
	}

	return &dashboard.MonthlyAttendanceResponse{
		OnTime:  data.OnTime,
		Late:    data.Late,
		Absent:  data.Absent,
		Records: data.Records,
	}, nil
}

// GetDailyAttendanceStats returns attendance stats with percentages for a specific day (1 query)
func (s *DashboardServiceImpl) GetDailyAttendanceStats(ctx context.Context, date string) (*dashboard.AttendanceStatsResponse, error) {
	companyID, err := s.getCompanyID(ctx)
	if err != nil {
		return nil, err
	}

	d := parseDate(date)

	stats, err := s.GetAttendanceStatsByDay(ctx, companyID, d)
	if err != nil {
		return nil, err
	}

	total := stats.OnTime + stats.Late + stats.Absent
	var onTimePercent, latePercent, absentPercent float64
	if total > 0 {
		onTimePercent = float64(stats.OnTime) / float64(total) * 100
		latePercent = float64(stats.Late) / float64(total) * 100
		absentPercent = float64(stats.Absent) / float64(total) * 100
	}

	return &dashboard.AttendanceStatsResponse{
		OnTime:        stats.OnTime,
		Late:          stats.Late,
		Absent:        stats.Absent,
		OnTimePercent: onTimePercent,
		LatePercent:   latePercent,
		AbsentPercent: absentPercent,
	}, nil
}
