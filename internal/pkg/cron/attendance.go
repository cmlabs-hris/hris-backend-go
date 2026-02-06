package cron

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/attendance"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/employee"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/master/branch"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/notification"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/schedule"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
)

type AttendanceJobs struct {
	attendanceRepo   attendance.AttendanceRepository
	employeeRepo     employee.EmployeeRepository
	scheduleRepo     schedule.WorkScheduleRepository
	scheduleTimeRepo schedule.WorkScheduleTimeRepository
	branchRepo       branch.BranchRepository
	notificationSvc  notification.Service
	db               *database.DB
}

func NewAttendanceJobs(
	attendanceRepo attendance.AttendanceRepository,
	employeeRepo employee.EmployeeRepository,
	scheduleRepo schedule.WorkScheduleRepository,
	scheduleTimeRepo schedule.WorkScheduleTimeRepository,
	branchRepo branch.BranchRepository,
	notificationSvc notification.Service,
	db *database.DB,
) *AttendanceJobs {
	return &AttendanceJobs{
		attendanceRepo:   attendanceRepo,
		employeeRepo:     employeeRepo,
		scheduleRepo:     scheduleRepo,
		scheduleTimeRepo: scheduleTimeRepo,
		branchRepo:       branchRepo,
		notificationSvc:  notificationSvc,
		db:               db,
	}
}

func (j *AttendanceJobs) RegisterJobs(scheduler *Scheduler) {
	scheduler.AddJob("auto_close_stale_attendances", 1*time.Hour, j.AutoCloseStaleAttendances)
	scheduler.AddJob("mark_absent_employees", 1*time.Hour, j.MarkAbsentEmployees)
}

func (j *AttendanceJobs) AutoCloseStaleAttendances(ctx context.Context) error {
	// Only run at midnight (00:00-00:59 UTC)
	if time.Now().UTC().Hour() != 0 {
		return nil
	}

	slog.Info("Cron: Starting auto-close stale attendances job")

	staleSessions, err := j.attendanceRepo.GetStaleOpenSessions(ctx, 2)
	if err != nil {
		return fmt.Errorf("failed to get stale sessions: %w", err)
	}

	if len(staleSessions) == 0 {
		slog.Info("Cron: No stale attendances found")
		return nil
	}

	closedCount := 0
	for _, session := range staleSessions {
		// Parse schedule metadata from rejection_reason field
		scheduleData := *session.RejectionReason
		parts := strings.Split(scheduleData, "|")
		var clockOutTimeStr, timezone string
		var isNextDay bool

		for _, part := range parts {
			kv := strings.Split(part, "=")
			if len(kv) == 2 {
				switch kv[0] {
				case "schedule_clock_out":
					clockOutTimeStr = kv[1]
				case "next_day":
					isNextDay = kv[1] == "true"
				case "timezone":
					timezone = kv[1]
				}
			}
		}

		// Load location
		loc, err := time.LoadLocation(timezone)
		if err != nil {
			loc = time.UTC
		}

		// Parse scheduled clock out time
		clockOutTime, _ := time.Parse("15:04:05", clockOutTimeStr)

		// Construct scheduled clock out timestamp
		scheduledOut := time.Date(
			session.Date.Year(), session.Date.Month(), session.Date.Day(),
			clockOutTime.Hour(), clockOutTime.Minute(), clockOutTime.Second(), 0,
			loc,
		)

		if isNextDay {
			scheduledOut = scheduledOut.Add(24 * time.Hour)
		}

		scheduledOutUTC := scheduledOut.UTC()

		// Calculate work hours
		workDuration := scheduledOutUTC.Sub(*session.ClockIn)
		workHoursMins := int(workDuration.Minutes())

		// Update attendance
		session.ClockOut = &scheduledOutUTC
		session.WorkHoursInMinutes = &workHoursMins
		session.Status = "auto_closed"
		reason := "Auto-closed: No clock-out detected within 2 hours of scheduled end time. Please contact your manager if this is incorrect."
		session.RejectionReason = &reason

		// Calculate overtime if applicable
		if session.WorkScheduleTimeID != nil {
			schedTime, err := j.scheduleTimeRepo.GetByID(ctx, *session.WorkScheduleTimeID, session.CompanyID)
			if err == nil {
				expectedDuration := schedTime.ClockOutTime.Sub(schedTime.ClockInTime)
				expectedMins := int(expectedDuration.Minutes())

				if workHoursMins > expectedMins {
					overtime := workHoursMins - expectedMins
					session.OvertimeMinutes = &overtime
				}
			}
		}

		if err := j.attendanceRepo.Update(ctx, session); err != nil {
			slog.Error("Cron: Failed to auto-close attendance",
				"attendance_id", session.ID,
				"employee_id", session.EmployeeID,
				"error", err)
			continue
		}

		// Notify employee
		if j.notificationSvc != nil {
			emp, _ := j.employeeRepo.GetByID(ctx, session.EmployeeID)
			if emp.UserID != nil {
				_ = j.notificationSvc.QueueNotification(ctx, notification.CreateNotificationRequest{
					CompanyID:   session.CompanyID,
					RecipientID: *emp.UserID,
					Type:        notification.TypeAttendanceAutoClosed,
					Title:       "Attendance Auto-Closed",
					Message:     fmt.Sprintf("Your attendance for %s was automatically closed", session.Date.Format("2006-01-02")),
					Data: map[string]interface{}{
						"attendance_id": session.ID,
						"date":          session.Date.Format("2006-01-02"),
						"reason":        reason,
					},
				})
			}

			// Notify managers
			managers, _ := j.employeeRepo.GetManagersByCompanyID(ctx, session.CompanyID)
			for _, manager := range managers {
				if manager.UserID != nil {
					_ = j.notificationSvc.QueueNotification(ctx, notification.CreateNotificationRequest{
						CompanyID:   session.CompanyID,
						RecipientID: *manager.UserID,
						SenderID:    emp.UserID,
						Type:        notification.TypeAttendanceAutoClosed,
						Title:       "Employee Attendance Auto-Closed",
						Message:     fmt.Sprintf("%s's attendance for %s was auto-closed", emp.FullName, session.Date.Format("2006-01-02")),
						Data: map[string]interface{}{
							"employee_id":   session.EmployeeID,
							"attendance_id": session.ID,
							"date":          session.Date.Format("2006-01-02"),
						},
					})
				}
			}
		}

		closedCount++
	}

	slog.Info("Cron: Auto-closed stale attendances", "count", closedCount)
	return nil
}

func (j *AttendanceJobs) MarkAbsentEmployees(ctx context.Context) error {
	// Only run at midnight (00:00-00:59 UTC)
	if time.Now().UTC().Hour() != 0 {
		return nil
	}

	slog.Info("Cron: Starting mark absent employees job")

	// Get all distinct company IDs from active employees
	rows, err := j.db.Pool.Query(ctx, `
		SELECT DISTINCT company_id FROM employees 
		WHERE employment_status = 'active' AND deleted_at IS NULL
	`)
	if err != nil {
		return fmt.Errorf("failed to get companies: %w", err)
	}
	defer rows.Close()

	var companyIDs []string
	for rows.Next() {
		var companyID string
		if err := rows.Scan(&companyID); err != nil {
			continue
		}
		companyIDs = append(companyIDs, companyID)
	}

	totalAbsent := 0

	for _, companyID := range companyIDs {
		// Get all active employees in this company
		employees, err := j.employeeRepo.GetActiveByCompanyID(ctx, companyID)
		if err != nil {
			slog.Error("Cron: Failed to get employees", "company_id", companyID, "error", err)
			continue
		}

		var absences []attendance.Attendance

		for _, emp := range employees {
			// Get timezone for employee
			timezone, err := j.branchRepo.GetTimezoneByEmployeeID(ctx, emp.ID, companyID)
			if err != nil {
				continue
			}

			loc, _ := time.LoadLocation(timezone)

			// Calculate yesterday in employee's timezone
			nowLocal := time.Now().In(loc)
			yesterdayLocal := nowLocal.AddDate(0, 0, -1)
			yesterdayStr := yesterdayLocal.Format("2006-01-02")

			// Check if employee has schedule for yesterday
			activeSchedule, err := j.scheduleRepo.GetActiveSchedule(ctx, emp.ID, yesterdayLocal, companyID)
			if err != nil || activeSchedule == nil {
				// No schedule yesterday, skip
				continue
			}

			// Check if already has attendance record
			hasAttendance, _ := j.attendanceRepo.HasCheckedInToday(ctx, emp.ID, yesterdayStr, companyID)
			if hasAttendance {
				// Already has record (either clocked in or marked absent), skip
				continue
			}

			// Create absence record
			zero := 0
			absenceRecord := attendance.Attendance{
				EmployeeID:         emp.ID,
				CompanyID:          companyID,
				Date:               yesterdayLocal.Truncate(24 * time.Hour),
				WorkScheduleTimeID: &activeSchedule.TimeID,
				Status:             "absent",
				WorkHoursInMinutes: &zero,
				ClockIn:            nil,
				ClockOut:           nil,
			}

			absences = append(absences, absenceRecord)
		}

		// Bulk insert absences
		if len(absences) > 0 {
			if err := j.attendanceRepo.BulkCreateAbsences(ctx, absences); err != nil {
				slog.Error("Cron: Failed to bulk create absences", "company_id", companyID, "error", err)
				continue
			}

			totalAbsent += len(absences)

			// Notify managers about absences
			if j.notificationSvc != nil {
				managers, _ := j.employeeRepo.GetManagersByCompanyID(ctx, companyID)
				for _, manager := range managers {
					if manager.UserID != nil {
						_ = j.notificationSvc.QueueNotification(ctx, notification.CreateNotificationRequest{
							CompanyID:   companyID,
							RecipientID: *manager.UserID,
							Type:        notification.TypeAttendanceMarkedAbsent,
							Title:       "Employees Marked Absent",
							Message:     fmt.Sprintf("%d employees were marked absent for yesterday", len(absences)),
							Data: map[string]interface{}{
								"count": len(absences),
								"date":  time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
							},
						})
					}
				}
			}
		}
	}

	slog.Info("Cron: Marked absent employees", "count", totalAbsent)
	return nil
}
