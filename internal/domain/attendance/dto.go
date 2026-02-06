package attendance

import (
	"mime/multipart"
	"strings"

	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/validator"
)

// ========================================
// ATTENDANCE DTOs
// ========================================

type ClockInRequest struct {
	EmployeeID    string                `json:"employee_id"`
	Latitude      float64               `json:"latitude"`
	Longitude     float64               `json:"longitude"`
	ProofPhotoURL *string               `json:"-"`
	File          multipart.File        `json:"-"`
	FileHeader    *multipart.FileHeader `json:"-"`
}

func (r *ClockInRequest) Validate() error {
	var errs validator.ValidationErrors

	if validator.IsEmpty(r.EmployeeID) {
		errs = append(errs, validator.ValidationError{
			Field:   "employee_id",
			Message: "employee_id is required",
		})
	}

	if r.Latitude < -90 || r.Latitude > 90 {
		errs = append(errs, validator.ValidationError{
			Field:   "latitude",
			Message: "latitude must be between -90 and 90",
		})
	}

	if r.Longitude < -180 || r.Longitude > 180 {
		errs = append(errs, validator.ValidationError{
			Field:   "longitude",
			Message: "longitude must be between -180 and 180",
		})
	}

	filename := r.FileHeader.Filename
	ext := strings.ToLower(filename[strings.LastIndex(filename, "."):])
	if r.FileHeader == nil {
		errs = append(errs, validator.ValidationError{
			Field:   "file",
			Message: "attendance proof photo is required",
		})
	} else if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		// Validate image format
		errs = append(errs, validator.ValidationError{
			Field:   "file",
			Message: "invalid file type: only jpg, jpeg, png allowed",
		})
	} else if r.FileHeader.Size > 10<<20 { // 10MB
		errs = append(errs, validator.ValidationError{
			Field:   "file",
			Message: "attendance proof photo size must not exceed 10MB",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type ClockOutRequest struct {
	EmployeeID    string                `json:"employee_id"`
	Latitude      float64               `json:"latitude"`
	Longitude     float64               `json:"longitude"`
	ProofPhotoURL *string               `json:"-"`
	File          multipart.File        `json:"-"`
	FileHeader    *multipart.FileHeader `json:"-"`
}

func (r *ClockOutRequest) Validate() error {
	var errs validator.ValidationErrors

	if validator.IsEmpty(r.EmployeeID) {
		errs = append(errs, validator.ValidationError{
			Field:   "employee_id",
			Message: "employee_id is required",
		})
	}

	if r.Latitude < -90 || r.Latitude > 90 {
		errs = append(errs, validator.ValidationError{
			Field:   "latitude",
			Message: "latitude must be between -90 and 90",
		})
	}

	if r.Longitude < -180 || r.Longitude > 180 {
		errs = append(errs, validator.ValidationError{
			Field:   "longitude",
			Message: "longitude must be between -180 and 180",
		})
	}

	filename := r.FileHeader.Filename
	ext := strings.ToLower(filename[strings.LastIndex(filename, "."):])
	if r.FileHeader == nil {
		errs = append(errs, validator.ValidationError{
			Field:   "file",
			Message: "attendance proof photo is required",
		})
	} else if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		// Validate image format
		errs = append(errs, validator.ValidationError{
			Field:   "file",
			Message: "invalid file type: only jpg, jpeg, png allowed",
		})
	} else if r.FileHeader.Size > 10<<20 { // 10MB
		errs = append(errs, validator.ValidationError{
			Field:   "file",
			Message: "attendance proof photo size must not exceed 10MB",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type AttendanceResponse struct {
	ID                string   `json:"id"`
	EmployeeID        string   `json:"employee_id"`
	EmployeeName      string   `json:"employee_name"`
	EmployeePosition  *string  `json:"employee_position,omitempty"`
	Date              string   `json:"date"`
	ClockInTime       *string  `json:"clock_in_time,omitempty"`
	ClockOutTime      *string  `json:"clock_out_time,omitempty"`
	ClockInLatitude   *float64 `json:"clock_in_latitude,omitempty"`
	ClockInLongitude  *float64 `json:"clock_in_longitude,omitempty"`
	ClockOutLatitude  *float64 `json:"clock_out_latitude,omitempty"`
	ClockOutLongitude *float64 `json:"clock_out_longitude,omitempty"`
	ClockInProofURL   *string  `json:"clock_in_proof_url,omitempty"`
	ClockOutProofURL  *string  `json:"clock_out_proof_url,omitempty"`
	WorkingHours      *float64 `json:"working_hours,omitempty"`
	Status            string   `json:"status"`
	IsLate            *bool    `json:"is_late,omitempty"`
	IsEarlyLeave      *bool    `json:"is_early_leave,omitempty"`
	LateMinutes       *int     `json:"late_minutes,omitempty"`
	EarlyLeaveMinutes *int     `json:"early_leave_minutes,omitempty"`
	CreatedAt         string   `json:"created_at"`
	UpdatedAt         string   `json:"updated_at"`
}

type AttendanceFilter struct {
	// Search & Filter
	EmployeeID   *string `json:"employee_id,omitempty"`
	EmployeeName *string `json:"employee_name,omitempty"`
	Date         *string `json:"date,omitempty"`       // YYYY-MM-DD
	StartDate    *string `json:"start_date,omitempty"` // YYYY-MM-DD
	EndDate      *string `json:"end_date,omitempty"`   // YYYY-MM-DD
	Status       *string `json:"status,omitempty"`

	// Pagination
	Page  int `json:"page"`
	Limit int `json:"limit"`

	// Sorting
	SortBy    string `json:"sort_by"`    // date, employee_name, clock_in_time, clock_out_time, status
	SortOrder string `json:"sort_order"` // asc, desc
}

func (f *AttendanceFilter) Validate() error {
	var errs validator.ValidationErrors

	// Page validation
	if f.Page < 0 {
		errs = append(errs, validator.ValidationError{
			Field:   "page",
			Message: "page must be a positive number",
		})
	}
	if f.Page == 0 {
		f.Page = 1 // Default page
	}

	// Limit validation
	if f.Limit < 0 {
		errs = append(errs, validator.ValidationError{
			Field:   "limit",
			Message: "limit must be a positive number",
		})
	}
	if f.Limit == 0 {
		f.Limit = 20 // Default limit
	}
	if f.Limit > 100 {
		errs = append(errs, validator.ValidationError{
			Field:   "limit",
			Message: "limit must not exceed 100",
		})
	}

	// Status validation
	if f.Status != nil {
		validStatuses := []string{"present", "absent", "late", "on_leave", "holiday"}
		if !validator.IsInSlice(*f.Status, validStatuses) {
			errs = append(errs, validator.ValidationError{
				Field:   "status",
				Message: "status must be one of: present, absent, late, on_leave, holiday",
			})
		}
	}

	// Date validation
	if f.Date != nil && *f.Date != "" {
		if _, valid := validator.IsValidDate(*f.Date); !valid {
			errs = append(errs, validator.ValidationError{
				Field:   "date",
				Message: "date must be in YYYY-MM-DD format",
			})
		}
	}

	if f.StartDate != nil && *f.StartDate != "" {
		if _, valid := validator.IsValidDate(*f.StartDate); !valid {
			errs = append(errs, validator.ValidationError{
				Field:   "start_date",
				Message: "start_date must be in YYYY-MM-DD format",
			})
		}
	}

	if f.EndDate != nil && *f.EndDate != "" {
		if _, valid := validator.IsValidDate(*f.EndDate); !valid {
			errs = append(errs, validator.ValidationError{
				Field:   "end_date",
				Message: "end_date must be in YYYY-MM-DD format",
			})
		}
	}

	// Sort validation
	if f.SortBy != "" {
		validSortFields := []string{"date", "employee_name", "clock_in_time", "clock_out_time", "status"}
		if !validator.IsInSlice(f.SortBy, validSortFields) {
			errs = append(errs, validator.ValidationError{
				Field:   "sort_by",
				Message: "sort_by must be one of: date, employee_name, clock_in_time, clock_out_time, status",
			})
		}
	} else {
		f.SortBy = "date" // Default sort
	}

	if f.SortOrder != "" {
		validSortOrders := []string{"asc", "desc"}
		if !validator.IsInSlice(strings.ToLower(f.SortOrder), validSortOrders) {
			errs = append(errs, validator.ValidationError{
				Field:   "sort_order",
				Message: "sort_order must be one of: asc, desc",
			})
		}
	} else {
		f.SortOrder = "desc" // Default descending (newest first)
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type MyAttendanceFilter struct {
	// Search & Filter (no employee filters)
	Date      *string `json:"date,omitempty"`       // YYYY-MM-DD
	StartDate *string `json:"start_date,omitempty"` // YYYY-MM-DD
	EndDate   *string `json:"end_date,omitempty"`   // YYYY-MM-DD
	Status    *string `json:"status,omitempty"`

	// Pagination
	Page  int `json:"page"`
	Limit int `json:"limit"`

	// Sorting
	SortBy    string `json:"sort_by"`    // date, clock_in_time, clock_out_time, status
	SortOrder string `json:"sort_order"` // asc, desc
}

func (f *MyAttendanceFilter) Validate() error {
	var errs validator.ValidationErrors

	// Page validation
	if f.Page < 0 {
		errs = append(errs, validator.ValidationError{
			Field:   "page",
			Message: "page must be a positive number",
		})
	}
	if f.Page == 0 {
		f.Page = 1 // Default page
	}

	// Limit validation
	if f.Limit < 0 {
		errs = append(errs, validator.ValidationError{
			Field:   "limit",
			Message: "limit must be a positive number",
		})
	}
	if f.Limit == 0 {
		f.Limit = 20 // Default limit
	}
	if f.Limit > 100 {
		errs = append(errs, validator.ValidationError{
			Field:   "limit",
			Message: "limit must not exceed 100",
		})
	}

	// Status validation
	if f.Status != nil {
		validStatuses := []string{"present", "absent", "late", "on_leave", "holiday", "waiting_approval"}
		if !validator.IsInSlice(*f.Status, validStatuses) {
			errs = append(errs, validator.ValidationError{
				Field:   "status",
				Message: "status must be one of: present, absent, late, on_leave, holiday, waiting_approval",
			})
		}
	}

	// Date validation
	if f.Date != nil && *f.Date != "" {
		if _, valid := validator.IsValidDate(*f.Date); !valid {
			errs = append(errs, validator.ValidationError{
				Field:   "date",
				Message: "date must be in YYYY-MM-DD format",
			})
		}
	}

	if f.StartDate != nil && *f.StartDate != "" {
		if _, valid := validator.IsValidDate(*f.StartDate); !valid {
			errs = append(errs, validator.ValidationError{
				Field:   "start_date",
				Message: "start_date must be in YYYY-MM-DD format",
			})
		}
	}

	if f.EndDate != nil && *f.EndDate != "" {
		if _, valid := validator.IsValidDate(*f.EndDate); !valid {
			errs = append(errs, validator.ValidationError{
				Field:   "end_date",
				Message: "end_date must be in YYYY-MM-DD format",
			})
		}
	}

	// Sort validation (no employee_name for my attendance)
	if f.SortBy != "" {
		validSortFields := []string{"date", "clock_in_time", "clock_out_time", "status"}
		if !validator.IsInSlice(f.SortBy, validSortFields) {
			errs = append(errs, validator.ValidationError{
				Field:   "sort_by",
				Message: "sort_by must be one of: date, clock_in_time, clock_out_time, status",
			})
		}
	} else {
		f.SortBy = "date" // Default sort
	}

	if f.SortOrder != "" {
		validSortOrders := []string{"asc", "desc"}
		if !validator.IsInSlice(strings.ToLower(f.SortOrder), validSortOrders) {
			errs = append(errs, validator.ValidationError{
				Field:   "sort_order",
				Message: "sort_order must be one of: asc, desc",
			})
		}
	} else {
		f.SortOrder = "desc" // Default descending (newest first)
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type ListAttendanceResponse struct {
	TotalCount  int64                `json:"total_count"`
	Page        int                  `json:"page"`
	Limit       int                  `json:"limit"`
	TotalPages  int                  `json:"total_pages"`
	Showing     string               `json:"showing"`
	Attendances []AttendanceResponse `json:"attendances"`
}

// UpdateAttendanceRequest for admin/manager to update attendance records
// This allows fixing wrong attendance data, employee forgot to clock in/out, etc.
type UpdateAttendanceRequest struct {
	ID                string   `json:"-"`
	Date              *string  `json:"date,omitempty"`           // YYYY-MM-DD
	ClockInTime       *string  `json:"clock_in_time,omitempty"`  // HH:MM:SS or full datetime
	ClockOutTime      *string  `json:"clock_out_time,omitempty"` // HH:MM:SS or full datetime
	ClockInLatitude   *float64 `json:"clock_in_latitude,omitempty"`
	ClockInLongitude  *float64 `json:"clock_in_longitude,omitempty"`
	ClockOutLatitude  *float64 `json:"clock_out_latitude,omitempty"`
	ClockOutLongitude *float64 `json:"clock_out_longitude,omitempty"`
	Status            *string  `json:"status,omitempty"`
	LateMinutes       *int     `json:"late_minutes,omitempty"`
	EarlyLeaveMinutes *int     `json:"early_leave_minutes,omitempty"`
	OvertimeMinutes   *int     `json:"overtime_minutes,omitempty"`
}

func (r *UpdateAttendanceRequest) Validate() error {
	var errs validator.ValidationErrors

	if r.Date != nil && *r.Date != "" {
		if _, valid := validator.IsValidDate(*r.Date); !valid {
			errs = append(errs, validator.ValidationError{
				Field:   "date",
				Message: "date must be in YYYY-MM-DD format",
			})
		}
	}

	if r.Status != nil {
		validStatuses := []string{"present", "absent", "late", "on_leave", "holiday", "waiting_approval"}
		if !validator.IsInSlice(strings.ToLower(*r.Status), validStatuses) {
			errs = append(errs, validator.ValidationError{
				Field:   "status",
				Message: "status must be one of: present, absent, late, on_leave, holiday, waiting_approval",
			})
		}
	}

	if r.ClockInLatitude != nil && (*r.ClockInLatitude < -90 || *r.ClockInLatitude > 90) {
		errs = append(errs, validator.ValidationError{
			Field:   "clock_in_latitude",
			Message: "clock_in_latitude must be between -90 and 90",
		})
	}

	if r.ClockInLongitude != nil && (*r.ClockInLongitude < -180 || *r.ClockInLongitude > 180) {
		errs = append(errs, validator.ValidationError{
			Field:   "clock_in_longitude",
			Message: "clock_in_longitude must be between -180 and 180",
		})
	}

	if r.ClockOutLatitude != nil && (*r.ClockOutLatitude < -90 || *r.ClockOutLatitude > 90) {
		errs = append(errs, validator.ValidationError{
			Field:   "clock_out_latitude",
			Message: "clock_out_latitude must be between -90 and 90",
		})
	}

	if r.ClockOutLongitude != nil && (*r.ClockOutLongitude < -180 || *r.ClockOutLongitude > 180) {
		errs = append(errs, validator.ValidationError{
			Field:   "clock_out_longitude",
			Message: "clock_out_longitude must be between -180 and 180",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

// ApproveAttendanceRequest for approving attendance
type ApproveAttendanceRequest struct {
	ID    string  `json:"-"`
	Notes *string `json:"notes,omitempty"` // Optional approval notes
}

// RejectAttendanceRequest for rejecting attendance
type RejectAttendanceRequest struct {
	ID     string `json:"-"`
	Reason string `json:"reason"` // Required rejection reason
}

func (r *RejectAttendanceRequest) Validate() error {
	var errs validator.ValidationErrors

	if validator.IsEmpty(r.Reason) {
		errs = append(errs, validator.ValidationError{
			Field:   "reason",
			Message: "rejection reason is required",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

// ========================================
// ATTENDANCE STATUS DTOs
// ========================================

type AttendanceStatusResponse struct {
	HasScheduleToday bool                `json:"has_schedule_today"`
	ScheduleInfo     *ActiveScheduleInfo `json:"schedule_info,omitempty"`
	HasCheckedIn     bool                `json:"has_checked_in"`
	TodayAttendance  *AttendanceResponse `json:"today_attendance,omitempty"`
	HasOpenSession   bool                `json:"has_open_session"`
	OpenSessionDate  string              `json:"open_session_date,omitempty"`
	OpenSessionID    string              `json:"open_session_id,omitempty"`
	CanClockIn       bool                `json:"can_clock_in"`
	CanClockOut      bool                `json:"can_clock_out"`
	Message          string              `json:"message"`
}

type ActiveScheduleInfo struct {
	ScheduleName       string `json:"schedule_name"`
	ClockInTime        string `json:"clock_in_time"`
	ClockOutTime       string `json:"clock_out_time"`
	IsNextDayCheckout  bool   `json:"is_next_day_checkout"`
	LocationType       string `json:"location_type"`
	GracePeriodMinutes int    `json:"grace_period_minutes"`
}
