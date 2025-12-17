package report

import (
	"context"
	"fmt"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/report"
	"github.com/cmlabs-hris/hris-backend-go/internal/repository/postgresql"
	"github.com/go-chi/jwtauth/v5"
)

type ReportServiceImpl struct {
	reportRepo postgresql.ReportRepository
}

func NewReportService(reportRepo postgresql.ReportRepository) report.ReportService {
	return &ReportServiceImpl{
		reportRepo: reportRepo,
	}
}

// getCompanyIDFromContext extracts company_id from JWT claims
func (s *ReportServiceImpl) getCompanyIDFromContext(ctx context.Context) (string, error) {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to extract claims from context: %w", err)
	}

	companyID, ok := claims["company_id"].(string)
	if !ok || companyID == "" {
		return "", fmt.Errorf("company_id claim is missing or invalid")
	}

	return companyID, nil
}

// GenerateMonthlyAttendanceReport generates the monthly attendance report
func (s *ReportServiceImpl) GenerateMonthlyAttendanceReport(ctx context.Context, req report.MonthlyAttendanceReportRequest) (report.MonthlyAttendanceReport, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return report.MonthlyAttendanceReport{}, err
	}

	// Get company ID from context
	companyID, err := s.getCompanyIDFromContext(ctx)
	if err != nil {
		return report.MonthlyAttendanceReport{}, err
	}

	// Calculate period dates
	periodStart := time.Date(req.Year, time.Month(req.Month), 1, 0, 0, 0, 0, time.Local)
	periodEnd := periodStart.AddDate(0, 1, -1)

	// Get data from repository
	employees, err := s.reportRepo.GetMonthlyAttendanceReport(ctx, companyID, req.Month, req.Year)
	if err != nil {
		return report.MonthlyAttendanceReport{}, fmt.Errorf("failed to get attendance data: %w", err)
	}

	return report.MonthlyAttendanceReport{
		PeriodMonth: req.Month,
		PeriodYear:  req.Year,
		PeriodStart: periodStart.Format("2006-01-02"),
		PeriodEnd:   periodEnd.Format("2006-01-02"),
		GeneratedAt: time.Now().Format(time.RFC3339),
		Employees:   employees,
	}, nil
}

// GeneratePayrollSummaryReport generates the payroll summary report
func (s *ReportServiceImpl) GeneratePayrollSummaryReport(ctx context.Context, req report.PayrollSummaryReportRequest) (report.PayrollSummaryReport, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return report.PayrollSummaryReport{}, err
	}

	// Get company ID from context
	companyID, err := s.getCompanyIDFromContext(ctx)
	if err != nil {
		return report.PayrollSummaryReport{}, err
	}

	// Get data from repository
	rows, err := s.reportRepo.GetPayrollSummaryReport(ctx, companyID, req.Month, req.Year)
	if err != nil {
		return report.PayrollSummaryReport{}, fmt.Errorf("failed to get payroll data: %w", err)
	}

	// Calculate totals
	var totalGross, totalNet float64
	for _, row := range rows {
		totalGross += row.GrossSalary
		totalNet += row.NetSalary
	}

	return report.PayrollSummaryReport{
		PeriodMonth:      req.Month,
		PeriodYear:       req.Year,
		GeneratedAt:      time.Now().Format(time.RFC3339),
		TotalGrossPayout: totalGross,
		TotalNetPayout:   totalNet,
		TotalEmployees:   len(rows),
		Rows:             rows,
	}, nil
}

// GenerateLeaveBalanceReport generates the leave balance report
func (s *ReportServiceImpl) GenerateLeaveBalanceReport(ctx context.Context, req report.LeaveBalanceReportRequest) (report.LeaveBalanceReport, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return report.LeaveBalanceReport{}, err
	}

	// Get company ID from context
	companyID, err := s.getCompanyIDFromContext(ctx)
	if err != nil {
		return report.LeaveBalanceReport{}, err
	}

	// Get data from repository
	rows, err := s.reportRepo.GetLeaveBalanceReport(ctx, companyID, req.Year)
	if err != nil {
		return report.LeaveBalanceReport{}, fmt.Errorf("failed to get leave balance data: %w", err)
	}

	return report.LeaveBalanceReport{
		GeneratedAt: time.Now().Format(time.RFC3339),
		Year:        req.Year,
		Rows:        rows,
	}, nil
}

// GenerateNewHireReport generates the new hire report
func (s *ReportServiceImpl) GenerateNewHireReport(ctx context.Context, req report.NewHireReportRequest) (report.NewHireReport, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return report.NewHireReport{}, err
	}

	// Get company ID from context
	companyID, err := s.getCompanyIDFromContext(ctx)
	if err != nil {
		return report.NewHireReport{}, err
	}

	// Get data from repository
	rows, err := s.reportRepo.GetNewHireReport(ctx, companyID, req.StartDate, req.EndDate)
	if err != nil {
		return report.NewHireReport{}, fmt.Errorf("failed to get new hire data: %w", err)
	}

	return report.NewHireReport{
		GeneratedAt: time.Now().Format(time.RFC3339),
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Rows:        rows,
	}, nil
}
