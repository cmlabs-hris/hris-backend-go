package fixtures

import (
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/leave"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/master/branch"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/master/grade"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/master/position"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/schedule"
)

// ==========================================
// HELPER FUNCTIONS
// ==========================================

func boolPtr(b bool) *bool          { return &b }
func intPtr(i int) *int             { return &i }
func strPtr(s string) *string       { return &s }
func float64Ptr(f float64) *float64 { return &f }

// ==========================================
// SEEDED DATA RESULT
// ==========================================

// SeededDataIDs holds IDs of all seeded default data for a company
type SeededDataIDs struct {
	// Position IDs by name
	PositionIDs map[string]string // e.g., "Director" -> "uuid"

	// Grade IDs by name
	GradeIDs map[string]string // e.g., "Executive" -> "uuid"

	// Branch ID (headquarters)
	BranchID string

	// Leave Type IDs by code
	LeaveTypeIDs map[string]string // e.g., "ANNUAL" -> "uuid"

	// Work Schedule ID (default schedule for owner - Standard Office Hours)
	WorkScheduleID string

	// Work Schedule IDs by name (for multiple schedules)
	WorkScheduleIDs map[string]string // e.g., "Standard Office Hours" -> "uuid"

	// Work Schedule Time IDs by schedule name and day of week
	WorkScheduleTimeIDs map[string]map[int]string // e.g., "Standard Office Hours" -> {1: "uuid", ...}
}

// NewSeededDataIDs creates a new SeededDataIDs with initialized maps
func NewSeededDataIDs() *SeededDataIDs {
	return &SeededDataIDs{
		PositionIDs:         make(map[string]string),
		GradeIDs:            make(map[string]string),
		LeaveTypeIDs:        make(map[string]string),
		WorkScheduleIDs:     make(map[string]string),
		WorkScheduleTimeIDs: make(map[string]map[int]string),
	}
}

// GetOwnerDefaults returns default IDs suitable for creating the company owner employee
// Returns: positionID (Director), gradeID (Executive), branchID (Headquarters)
func (s *SeededDataIDs) GetOwnerDefaults() (positionID, gradeID, branchID, workScheduleID string) {
	positionID = s.PositionIDs["Director"]
	gradeID = s.GradeIDs["Executive"]
	branchID = s.BranchID
	workScheduleID = s.WorkScheduleID
	return
}

// ==========================================
// DEFAULT POSITIONS
// ==========================================

// GetDefaultPositions returns standard job positions for a new company
func GetDefaultPositions(companyID string) []position.Position {
	return []position.Position{
		{CompanyID: companyID, Name: "Director"},
		{CompanyID: companyID, Name: "Manager"},
		{CompanyID: companyID, Name: "Supervisor"},
		{CompanyID: companyID, Name: "Team Lead"},
		{CompanyID: companyID, Name: "Senior Staff"},
		{CompanyID: companyID, Name: "Staff"},
		{CompanyID: companyID, Name: "Junior Staff"},
		{CompanyID: companyID, Name: "Intern"},
	}
}

// ==========================================
// DEFAULT GRADES
// ==========================================

// GetDefaultGrades returns standard employee grades/levels for a new company
func GetDefaultGrades(companyID string) []grade.Grade {
	return []grade.Grade{
		{CompanyID: companyID, Name: "Executive"},
		{CompanyID: companyID, Name: "Senior"},
		{CompanyID: companyID, Name: "Mid-Level"},
		{CompanyID: companyID, Name: "Junior"},
		{CompanyID: companyID, Name: "Entry Level"},
		{CompanyID: companyID, Name: "Trainee"},
	}
}

// ==========================================
// DEFAULT BRANCHES
// ==========================================

// GetDefaultBranch returns a default headquarters branch for a new company
func GetDefaultBranch(companyID string, companyName string) branch.Branch {
	return branch.Branch{
		CompanyID: companyID,
		Name:      "Headquarters",
		Address:   strPtr(companyName + " - Main Office"),
		Timezone:  "Asia/Jakarta", // Default to WIB (Indonesian Western Time)
	}
}

// ==========================================
// DEFAULT LEAVE TYPES
// ==========================================

// GetDefaultLeaveTypes returns standard leave types based on Indonesian labor law
func GetDefaultLeaveTypes(companyID string) []leave.LeaveType {
	return []leave.LeaveType{
		// Annual Leave (Cuti Tahunan) - 12 days per year after 12 months employment
		{
			CompanyID:                   companyID,
			Name:                        "Cuti Tahunan",
			Code:                        strPtr("ANNUAL"),
			Description:                 strPtr("Annual leave entitlement as per Indonesian Labor Law (12 days/year after 12 months of service)"),
			Color:                       strPtr("#4CAF50"), // Green
			IsActive:                    boolPtr(true),
			RequiresApproval:            boolPtr(true),
			RequiresAttachment:          boolPtr(false),
			AttachmentRequiredAfterDays: nil,
			HasQuota:                    boolPtr(true),
			AccrualMethod:               strPtr("yearly"),
			DeductionType:               strPtr("working_days"),
			AllowHalfDay:                boolPtr(true),
			MaxDaysPerRequest:           intPtr(12),
			MinNoticeDays:               intPtr(3),
			MaxAdvanceDays:              intPtr(30),
			AllowBackdate:               boolPtr(false),
			BackdateMaxDays:             nil,
			AllowRollover:               boolPtr(true),
			MaxRolloverDays:             intPtr(6),
			RolloverExpiryMonth:         intPtr(3), // Expires end of March
			QuotaCalculationType:        "fixed",
			QuotaRules: leave.QuotaRules{
				Type:         "fixed",
				DefaultQuota: 12,
			},
		},

		// Sick Leave (Cuti Sakit)
		{
			CompanyID:                   companyID,
			Name:                        "Cuti Sakit",
			Code:                        strPtr("SICK"),
			Description:                 strPtr("Sick leave with doctor's certificate required for more than 1 day"),
			Color:                       strPtr("#F44336"), // Red
			IsActive:                    boolPtr(true),
			RequiresApproval:            boolPtr(true),
			RequiresAttachment:          boolPtr(true),
			AttachmentRequiredAfterDays: intPtr(1), // Requires attachment after 1 day
			HasQuota:                    boolPtr(false),
			AccrualMethod:               strPtr("none"),
			DeductionType:               strPtr("calendar_days"),
			AllowHalfDay:                boolPtr(false),
			MaxDaysPerRequest:           nil, // No limit
			MinNoticeDays:               intPtr(0),
			MaxAdvanceDays:              intPtr(0),
			AllowBackdate:               boolPtr(true),
			BackdateMaxDays:             intPtr(3),
			AllowRollover:               boolPtr(false),
			MaxRolloverDays:             nil,
			RolloverExpiryMonth:         nil,
			QuotaCalculationType:        "fixed",
			QuotaRules: leave.QuotaRules{
				Type:         "fixed",
				DefaultQuota: 0, // Unlimited sick leave
			},
		},

		// Marriage Leave (Cuti Menikah) - 3 days
		{
			CompanyID:                   companyID,
			Name:                        "Cuti Menikah",
			Code:                        strPtr("MARRIAGE"),
			Description:                 strPtr("Marriage leave for employee's own wedding (3 days)"),
			Color:                       strPtr("#E91E63"), // Pink
			IsActive:                    boolPtr(true),
			RequiresApproval:            boolPtr(true),
			RequiresAttachment:          boolPtr(true),
			AttachmentRequiredAfterDays: nil,
			HasQuota:                    boolPtr(true),
			AccrualMethod:               strPtr("none"),
			DeductionType:               strPtr("calendar_days"),
			AllowHalfDay:                boolPtr(false),
			MaxDaysPerRequest:           intPtr(3),
			MinNoticeDays:               intPtr(7),
			MaxAdvanceDays:              intPtr(30),
			AllowBackdate:               boolPtr(false),
			BackdateMaxDays:             nil,
			AllowRollover:               boolPtr(false),
			MaxRolloverDays:             nil,
			RolloverExpiryMonth:         nil,
			QuotaCalculationType:        "fixed",
			QuotaRules: leave.QuotaRules{
				Type:         "fixed",
				DefaultQuota: 3,
			},
		},

		// Maternity Leave (Cuti Melahirkan) - 3 months
		{
			CompanyID:                   companyID,
			Name:                        "Cuti Melahirkan",
			Code:                        strPtr("MATERNITY"),
			Description:                 strPtr("Maternity leave for female employees (3 months: 1.5 before + 1.5 after delivery)"),
			Color:                       strPtr("#9C27B0"), // Purple
			IsActive:                    boolPtr(true),
			RequiresApproval:            boolPtr(true),
			RequiresAttachment:          boolPtr(true),
			AttachmentRequiredAfterDays: nil,
			HasQuota:                    boolPtr(true),
			AccrualMethod:               strPtr("none"),
			DeductionType:               strPtr("calendar_days"),
			AllowHalfDay:                boolPtr(false),
			MaxDaysPerRequest:           intPtr(90),
			MinNoticeDays:               intPtr(14),
			MaxAdvanceDays:              intPtr(45),
			AllowBackdate:               boolPtr(false),
			BackdateMaxDays:             nil,
			AllowRollover:               boolPtr(false),
			MaxRolloverDays:             nil,
			RolloverExpiryMonth:         nil,
			QuotaCalculationType:        "fixed",
			QuotaRules: leave.QuotaRules{
				Type:         "fixed",
				DefaultQuota: 90,
			},
		},

		// Paternity Leave (Cuti Ayah) - 2 days
		{
			CompanyID:                   companyID,
			Name:                        "Cuti Ayah",
			Code:                        strPtr("PATERNITY"),
			Description:                 strPtr("Paternity leave for male employees when wife gives birth (2 days)"),
			Color:                       strPtr("#3F51B5"), // Indigo
			IsActive:                    boolPtr(true),
			RequiresApproval:            boolPtr(true),
			RequiresAttachment:          boolPtr(true),
			AttachmentRequiredAfterDays: nil,
			HasQuota:                    boolPtr(true),
			AccrualMethod:               strPtr("none"),
			DeductionType:               strPtr("calendar_days"),
			AllowHalfDay:                boolPtr(false),
			MaxDaysPerRequest:           intPtr(2),
			MinNoticeDays:               intPtr(0),
			MaxAdvanceDays:              intPtr(14),
			AllowBackdate:               boolPtr(true),
			BackdateMaxDays:             intPtr(7),
			AllowRollover:               boolPtr(false),
			MaxRolloverDays:             nil,
			RolloverExpiryMonth:         nil,
			QuotaCalculationType:        "fixed",
			QuotaRules: leave.QuotaRules{
				Type:         "fixed",
				DefaultQuota: 2,
			},
		},

		// Bereavement Leave (Cuti Duka) - 2 days (spouse/child/parent) or 1 day (extended family)
		{
			CompanyID:                   companyID,
			Name:                        "Cuti Duka Keluarga Inti",
			Code:                        strPtr("BEREAVEMENT_IMMEDIATE"),
			Description:                 strPtr("Bereavement leave for immediate family (spouse, child, parent, parent-in-law) - 2 days"),
			Color:                       strPtr("#607D8B"), // Blue Grey
			IsActive:                    boolPtr(true),
			RequiresApproval:            boolPtr(true),
			RequiresAttachment:          boolPtr(false),
			AttachmentRequiredAfterDays: nil,
			HasQuota:                    boolPtr(true),
			AccrualMethod:               strPtr("none"),
			DeductionType:               strPtr("calendar_days"),
			AllowHalfDay:                boolPtr(false),
			MaxDaysPerRequest:           intPtr(2),
			MinNoticeDays:               intPtr(0),
			MaxAdvanceDays:              intPtr(0),
			AllowBackdate:               boolPtr(true),
			BackdateMaxDays:             intPtr(3),
			AllowRollover:               boolPtr(false),
			MaxRolloverDays:             nil,
			RolloverExpiryMonth:         nil,
			QuotaCalculationType:        "fixed",
			QuotaRules: leave.QuotaRules{
				Type:         "fixed",
				DefaultQuota: 2,
			},
		},

		// Bereavement Leave Extended Family - 1 day
		{
			CompanyID:                   companyID,
			Name:                        "Cuti Duka Keluarga Lain",
			Code:                        strPtr("BEREAVEMENT_EXTENDED"),
			Description:                 strPtr("Bereavement leave for extended family members - 1 day"),
			Color:                       strPtr("#78909C"), // Blue Grey Lighten
			IsActive:                    boolPtr(true),
			RequiresApproval:            boolPtr(true),
			RequiresAttachment:          boolPtr(false),
			AttachmentRequiredAfterDays: nil,
			HasQuota:                    boolPtr(true),
			AccrualMethod:               strPtr("none"),
			DeductionType:               strPtr("calendar_days"),
			AllowHalfDay:                boolPtr(false),
			MaxDaysPerRequest:           intPtr(1),
			MinNoticeDays:               intPtr(0),
			MaxAdvanceDays:              intPtr(0),
			AllowBackdate:               boolPtr(true),
			BackdateMaxDays:             intPtr(3),
			AllowRollover:               boolPtr(false),
			MaxRolloverDays:             nil,
			RolloverExpiryMonth:         nil,
			QuotaCalculationType:        "fixed",
			QuotaRules: leave.QuotaRules{
				Type:         "fixed",
				DefaultQuota: 1,
			},
		},

		// Child Circumcision/Baptism Leave - 2 days
		{
			CompanyID:                   companyID,
			Name:                        "Cuti Khitanan/Pembaptisan Anak",
			Code:                        strPtr("CHILD_CEREMONY"),
			Description:                 strPtr("Leave for child's circumcision or baptism ceremony (2 days)"),
			Color:                       strPtr("#FF9800"), // Orange
			IsActive:                    boolPtr(true),
			RequiresApproval:            boolPtr(true),
			RequiresAttachment:          boolPtr(false),
			AttachmentRequiredAfterDays: nil,
			HasQuota:                    boolPtr(true),
			AccrualMethod:               strPtr("none"),
			DeductionType:               strPtr("calendar_days"),
			AllowHalfDay:                boolPtr(false),
			MaxDaysPerRequest:           intPtr(2),
			MinNoticeDays:               intPtr(3),
			MaxAdvanceDays:              intPtr(14),
			AllowBackdate:               boolPtr(false),
			BackdateMaxDays:             nil,
			AllowRollover:               boolPtr(false),
			MaxRolloverDays:             nil,
			RolloverExpiryMonth:         nil,
			QuotaCalculationType:        "fixed",
			QuotaRules: leave.QuotaRules{
				Type:         "fixed",
				DefaultQuota: 2,
			},
		},

		// Unpaid Leave (Cuti Tanpa Gaji)
		{
			CompanyID:                   companyID,
			Name:                        "Cuti Tanpa Gaji",
			Code:                        strPtr("UNPAID"),
			Description:                 strPtr("Unpaid leave for personal matters"),
			Color:                       strPtr("#795548"), // Brown
			IsActive:                    boolPtr(true),
			RequiresApproval:            boolPtr(true),
			RequiresAttachment:          boolPtr(false),
			AttachmentRequiredAfterDays: nil,
			HasQuota:                    boolPtr(false),
			AccrualMethod:               strPtr("none"),
			DeductionType:               strPtr("working_days"),
			AllowHalfDay:                boolPtr(true),
			MaxDaysPerRequest:           intPtr(30),
			MinNoticeDays:               intPtr(7),
			MaxAdvanceDays:              intPtr(60),
			AllowBackdate:               boolPtr(false),
			BackdateMaxDays:             nil,
			AllowRollover:               boolPtr(false),
			MaxRolloverDays:             nil,
			RolloverExpiryMonth:         nil,
			QuotaCalculationType:        "fixed",
			QuotaRules: leave.QuotaRules{
				Type:         "fixed",
				DefaultQuota: 0,
			},
		},
	}
}

// ==========================================
// DEFAULT WORK SCHEDULE
// ==========================================

// GetDefaultWorkSchedule returns a standard 9-5 work schedule for a new company
func GetDefaultWorkSchedule(companyID string) schedule.WorkSchedule {
	return schedule.WorkSchedule{
		CompanyID:          companyID,
		Name:               "Standard Office Hours",
		Type:               schedule.WorkArrangementWFO,
		GracePeriodMinutes: 15, // 15 minutes grace period
	}
}

// GetDefaultWorkScheduleTimes returns standard work hours (Mon-Fri 09:00-18:00)
func GetDefaultWorkScheduleTimes(workScheduleID string) []schedule.WorkScheduleTime {
	// Parse time for clock in/out
	clockIn := time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)     // 09:00
	breakStart := time.Date(0, 1, 1, 12, 0, 0, 0, time.UTC) // 12:00
	breakEnd := time.Date(0, 1, 1, 13, 0, 0, 0, time.UTC)   // 13:00
	clockOut := time.Date(0, 1, 1, 18, 0, 0, 0, time.UTC)   // 18:00

	times := make([]schedule.WorkScheduleTime, 0, 5)

	// Monday to Friday (1-5)
	for day := 1; day <= 5; day++ {
		times = append(times, schedule.WorkScheduleTime{
			WorkScheduleID:    workScheduleID,
			DayOfWeek:         day,
			ClockInTime:       clockIn,
			BreakStartTime:    &breakStart,
			BreakEndTime:      &breakEnd,
			ClockOutTime:      clockOut,
			IsNextDayCheckout: false,
			LocationType:      schedule.WorkArrangementWFO,
		})
	}

	return times
}

// ==========================================
// NIGHT/OVERNIGHT SHIFT SCHEDULE
// ==========================================

// GetNightShiftWorkSchedule returns an overnight/night shift schedule (22:00-06:00)
func GetNightShiftWorkSchedule(companyID string) schedule.WorkSchedule {
	return schedule.WorkSchedule{
		CompanyID:          companyID,
		Name:               "Night Shift",
		Type:               schedule.WorkArrangementWFO,
		GracePeriodMinutes: 15, // 15 minutes grace period
	}
}

// GetNightShiftWorkScheduleTimes returns night shift hours (Mon-Fri 22:00-06:00 next day)
func GetNightShiftWorkScheduleTimes(workScheduleID string) []schedule.WorkScheduleTime {
	// Parse time for clock in/out
	clockIn := time.Date(0, 1, 1, 22, 0, 0, 0, time.UTC)   // 22:00
	breakStart := time.Date(0, 1, 1, 1, 0, 0, 0, time.UTC) // 01:00 (next day)
	breakEnd := time.Date(0, 1, 1, 2, 0, 0, 0, time.UTC)   // 02:00 (next day)
	clockOut := time.Date(0, 1, 1, 6, 0, 0, 0, time.UTC)   // 06:00 (next day)

	times := make([]schedule.WorkScheduleTime, 0, 5)

	// Monday to Friday (1-5) - night shift starts on these days
	for day := 1; day <= 5; day++ {
		times = append(times, schedule.WorkScheduleTime{
			WorkScheduleID:    workScheduleID,
			DayOfWeek:         day,
			ClockInTime:       clockIn,
			BreakStartTime:    &breakStart,
			BreakEndTime:      &breakEnd,
			ClockOutTime:      clockOut,
			IsNextDayCheckout: true, // Clock out is on the next day
			LocationType:      schedule.WorkArrangementWFO,
		})
	}

	return times
}

// ==========================================
// AFTERNOON/SECOND SHIFT SCHEDULE
// ==========================================

// GetAfternoonShiftWorkSchedule returns an afternoon/second shift schedule (14:00-22:00)
func GetAfternoonShiftWorkSchedule(companyID string) schedule.WorkSchedule {
	return schedule.WorkSchedule{
		CompanyID:          companyID,
		Name:               "Afternoon Shift",
		Type:               schedule.WorkArrangementWFO,
		GracePeriodMinutes: 15, // 15 minutes grace period
	}
}

// GetAfternoonShiftWorkScheduleTimes returns afternoon shift hours (Mon-Fri 14:00-22:00)
func GetAfternoonShiftWorkScheduleTimes(workScheduleID string) []schedule.WorkScheduleTime {
	// Parse time for clock in/out
	clockIn := time.Date(0, 1, 1, 14, 0, 0, 0, time.UTC)    // 14:00
	breakStart := time.Date(0, 1, 1, 18, 0, 0, 0, time.UTC) // 18:00
	breakEnd := time.Date(0, 1, 1, 19, 0, 0, 0, time.UTC)   // 19:00
	clockOut := time.Date(0, 1, 1, 22, 0, 0, 0, time.UTC)   // 22:00

	times := make([]schedule.WorkScheduleTime, 0, 5)

	// Monday to Friday (1-5)
	for day := 1; day <= 5; day++ {
		times = append(times, schedule.WorkScheduleTime{
			WorkScheduleID:    workScheduleID,
			DayOfWeek:         day,
			ClockInTime:       clockIn,
			BreakStartTime:    &breakStart,
			BreakEndTime:      &breakEnd,
			ClockOutTime:      clockOut,
			IsNextDayCheckout: false,
			LocationType:      schedule.WorkArrangementWFO,
		})
	}

	return times
}

// ==========================================
// FLEXIBLE/WFA SCHEDULE
// ==========================================

// GetFlexibleWorkSchedule returns a flexible/WFA work schedule
func GetFlexibleWorkSchedule(companyID string) schedule.WorkSchedule {
	return schedule.WorkSchedule{
		CompanyID:          companyID,
		Name:               "Flexible Hours (WFA)",
		Type:               schedule.WorkArrangementWFA,
		GracePeriodMinutes: 30, // More lenient grace period for flexible work
	}
}

// GetFlexibleWorkScheduleTimes returns flexible work hours (Mon-Fri 08:00-17:00)
func GetFlexibleWorkScheduleTimes(workScheduleID string) []schedule.WorkScheduleTime {
	// Parse time for clock in/out - flexible hours with wider window
	clockIn := time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC)     // 08:00
	breakStart := time.Date(0, 1, 1, 12, 0, 0, 0, time.UTC) // 12:00
	breakEnd := time.Date(0, 1, 1, 13, 0, 0, 0, time.UTC)   // 13:00
	clockOut := time.Date(0, 1, 1, 17, 0, 0, 0, time.UTC)   // 17:00

	times := make([]schedule.WorkScheduleTime, 0, 5)

	// Monday to Friday (1-5)
	for day := 1; day <= 5; day++ {
		times = append(times, schedule.WorkScheduleTime{
			WorkScheduleID:    workScheduleID,
			DayOfWeek:         day,
			ClockInTime:       clockIn,
			BreakStartTime:    &breakStart,
			BreakEndTime:      &breakEnd,
			ClockOutTime:      clockOut,
			IsNextDayCheckout: false,
			LocationType:      schedule.WorkArrangementWFA,
		})
	}

	return times
}

// ==========================================
// ALL DEFAULT WORK SCHEDULES
// ==========================================

// WorkScheduleDefinition holds a schedule and its time generator
type WorkScheduleDefinition struct {
	Schedule    schedule.WorkSchedule
	TimesGetter func(workScheduleID string) []schedule.WorkScheduleTime
}

// GetAllDefaultWorkSchedules returns all default work schedules for a new company
func GetAllDefaultWorkSchedules(companyID string) []WorkScheduleDefinition {
	return []WorkScheduleDefinition{
		{
			Schedule:    GetDefaultWorkSchedule(companyID),
			TimesGetter: GetDefaultWorkScheduleTimes,
		},
		{
			Schedule:    GetNightShiftWorkSchedule(companyID),
			TimesGetter: GetNightShiftWorkScheduleTimes,
		},
		{
			Schedule:    GetAfternoonShiftWorkSchedule(companyID),
			TimesGetter: GetAfternoonShiftWorkScheduleTimes,
		},
		{
			Schedule:    GetFlexibleWorkSchedule(companyID),
			TimesGetter: GetFlexibleWorkScheduleTimes,
		},
	}
}
