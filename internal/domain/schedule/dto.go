package schedule

import (
	"strings"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/validator"
)

type CreateWorkScheduleRequest struct {
	Name               string `json:"name"`
	Type               string `json:"type"`
	GracePeriodMinutes *int   `json:"grace_period_minutes"`
}

func (r *CreateWorkScheduleRequest) Validate() error {
	var errs validator.ValidationErrors

	if validator.IsEmpty(r.Name) {
		errs = append(errs, validator.ValidationError{
			Field:   "name",
			Message: "name is required",
		})
	}
	if validator.IsEmpty(r.Type) {
		errs = append(errs, validator.ValidationError{
			Field:   "type",
			Message: "type is required",
		})
	}
	if !validator.IsInSlice(r.Type, WorkArrangementValues) {
		errs = append(errs, validator.ValidationError{
			Field:   "type",
			Message: "type must be one of: " + strings.Join(WorkArrangementValues, ", "),
		})
	}
	if r.GracePeriodMinutes == nil {
		errs = append(errs, validator.ValidationError{
			Field:   "grace_period_minutes",
			Message: "grace_period_minutes is required",
		})
	}
	if r.GracePeriodMinutes != nil && *r.GracePeriodMinutes < 0 {
		errs = append(errs, validator.ValidationError{
			Field:   "grace_period_minutes",
			Message: "grace_period_minutes must be a non-negative number",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type WorkScheduleResponse struct {
	ID                 string                         `json:"id"`
	CompanyID          string                         `json:"company_id"`
	Name               string                         `json:"name"`
	Type               string                         `json:"type"`
	GracePeriodMinutes int                            `json:"grace_period_minutes"`
	Times              []WorkScheduleTimeResponse     `json:"times,omitempty"`
	Locations          []WorkScheduleLocationResponse `json:"locations,omitempty"`
	CreatedAt          string                         `json:"created_at"`
	UpdatedAt          string                         `json:"updated_at"`
	DeletedAt          *string                        `json:"deleted_at,omitempty"`
}

// ListWorkScheduleResponse - Enhanced with pagination metadata
type ListWorkScheduleResponse struct {
	TotalCount    int64                  `json:"total_count"`
	Page          int                    `json:"page"`
	Limit         int                    `json:"limit"`
	TotalPages    int                    `json:"total_pages"` // ← NEW
	Showing       string                 `json:"showing"`     // ← NEW: "21-40 of 150 results"
	WorkSchedules []WorkScheduleResponse `json:"work_schedules"`
}

type WorkScheduleFilter struct {
	// Search & Filter
	Name      *string `json:"name,omitempty"`       // Search by name
	Type      *string `json:"type,omitempty"`       // Filter by work arrangement type
	CompanyID *string `json:"company_id,omitempty"` // Filter by company ID

	// Pagination
	Page  int `json:"page"`
	Limit int `json:"limit"`

	// Sorting
	SortBy    string `json:"sort_by"`    // name, type, created_at
	SortOrder string `json:"sort_order"` // asc, desc
}

func (f *WorkScheduleFilter) Validate() error {
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

	// Sort validation
	if f.SortBy != "" {
		validSortFields := []string{"name", "type", "created_at"}
		if !validator.IsInSlice(f.SortBy, validSortFields) {
			errs = append(errs, validator.ValidationError{
				Field:   "sort_by",
				Message: "sort_by must be one of: name, type, created_at",
			})
		}
	} else {
		f.SortBy = "name" // Default sort
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
		f.SortOrder = "asc" // Default ascending
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type CreateWorkScheduleTimeRequest struct {
	WorkScheduleID    string  `json:"work_schedule_id"`
	DayOfWeek         *int    `json:"day_of_week"`
	ClockInTime       string  `json:"clock_in_time"`              // HH:MM format
	BreakStartTime    *string `json:"break_start_time,omitempty"` // HH:MM format, optional
	BreakEndTime      *string `json:"break_end_time,omitempty"`   // HH:MM format, optional
	ClockOutTime      string  `json:"clock_out_time"`             // HH:MM format
	IsNextDayCheckout *bool   `json:"is_next_day_checkout"`
	LocationType      string  `json:"location_type"`
}

func (r *CreateWorkScheduleTimeRequest) Validate() error {
	var errs validator.ValidationErrors

	if validator.IsEmpty(r.WorkScheduleID) {
		errs = append(errs, validator.ValidationError{
			Field:   "work_schedule_id",
			Message: "work_schedule_id is required",
		})
	}
	if r.DayOfWeek == nil {
		errs = append(errs, validator.ValidationError{
			Field:   "day_of_week",
			Message: "day_of_week is required",
		})
	}
	if r.DayOfWeek != nil && (*r.DayOfWeek < 1 || *r.DayOfWeek > 7) {
		errs = append(errs, validator.ValidationError{
			Field:   "day_of_week",
			Message: "day_of_week must be between 1 (Monday) and 7 (Sunday)",
		})
	}

	// Validate ClockInTime - required
	if validator.IsEmpty(r.ClockInTime) {
		errs = append(errs, validator.ValidationError{
			Field:   "clock_in_time",
			Message: "clock_in_time is required",
		})
	} else if _, valid := validator.IsValidTime(r.ClockInTime); !valid {
		errs = append(errs, validator.ValidationError{
			Field:   "clock_in_time",
			Message: "clock_in_time must be a valid time in HH:MM format",
		})
	}

	// Validate ClockOutTime - required
	if validator.IsEmpty(r.ClockOutTime) {
		errs = append(errs, validator.ValidationError{
			Field:   "clock_out_time",
			Message: "clock_out_time is required",
		})
	} else if _, valid := validator.IsValidTime(r.ClockOutTime); !valid {
		errs = append(errs, validator.ValidationError{
			Field:   "clock_out_time",
			Message: "clock_out_time must be a valid time in HH:MM format",
		})
	}

	if r.IsNextDayCheckout == nil {
		errs = append(errs, validator.ValidationError{
			Field:   "is_next_day_checkout",
			Message: "is_next_day_checkout is required",
		})
	}

	// Validate ClockInTime and ClockOutTime based on is_next_day_checkout
	if r.IsNextDayCheckout != nil && !*r.IsNextDayCheckout && r.ClockInTime > r.ClockOutTime {
		errs = append(errs, validator.ValidationError{
			Field:   "clock_in_time",
			Message: "clock_out_time must not be before clock_in_time when is_next_day_checkout is false",
		})
	}
	if r.IsNextDayCheckout != nil && *r.IsNextDayCheckout && r.ClockInTime < r.ClockOutTime {
		errs = append(errs, validator.ValidationError{
			Field:   "clock_in_time",
			Message: "clock_out_time must be before clock_in_time when is_next_day_checkout is true",
		})
	}

	// Validate break times (both must be provided or neither)
	if (r.BreakStartTime != nil && r.BreakEndTime == nil) || (r.BreakStartTime == nil && r.BreakEndTime != nil) {
		errs = append(errs, validator.ValidationError{
			Field:   "break_times",
			Message: "both break_start_time and break_end_time must be provided or neither",
		})
	} else if r.BreakStartTime != nil && r.BreakEndTime != nil {
		_, validStart := validator.IsValidTime(*r.BreakStartTime)
		_, validEnd := validator.IsValidTime(*r.BreakEndTime)

		if !validStart {
			errs = append(errs, validator.ValidationError{
				Field:   "break_start_time",
				Message: "break_start_time must be a valid time in HH:MM format",
			})
		}
		if !validEnd {
			errs = append(errs, validator.ValidationError{
				Field:   "break_end_time",
				Message: "break_end_time must be a valid time in HH:MM format",
			})
		}

		if validStart && validEnd {
			// Validate break times based on shift type
			if r.IsNextDayCheckout != nil && *r.IsNextDayCheckout {
				// ===== OVERNIGHT SHIFT VALIDATION =====
				// Example: Clock In 23:00 (Day 1) → Clock Out 07:00 (Day 2)
				// Valid break scenarios:
				// 1. Break before midnight: 23:30 - 23:59
				// 2. Break after midnight: 01:00 - 02:00
				// 3. Break spanning midnight: 23:30 - 00:30

				breakStartAfterClockIn := *r.BreakStartTime >= r.ClockInTime
				breakEndBeforeClockOut := *r.BreakEndTime <= r.ClockOutTime

				// Check if break is within the shift window
				// At least one of these must be true:
				// - Break starts after clock_in (same day portion)
				// - Break ends before clock_out (next day portion)
				validBreakWindow := breakStartAfterClockIn || breakEndBeforeClockOut

				if !validBreakWindow {
					errs = append(errs, validator.ValidationError{
						Field:   "break_times",
						Message: "break must be within shift window (after clock_in or before clock_out)",
					})
				}

				// Validate break sequence based on whether it spans midnight
				if *r.BreakStartTime < *r.BreakEndTime {
					// Case 1: Break doesn't span midnight
					// Examples: 23:30-23:59 OR 01:00-02:00
					// Both times are on same side of midnight - this is normal

					// If break is entirely before midnight (same day as clock_in)
					// it must start after clock_in
					if breakStartAfterClockIn && *r.BreakEndTime > "23:59" {
						errs = append(errs, validator.ValidationError{
							Field:   "break_end_time",
							Message: "break_end_time cannot extend past midnight if break_start_time is before midnight",
						})
					}
				} else {
					// Case 2: Break spans midnight (break_start > break_end)
					// Example: 23:30 - 00:30
					// This means break starts on Day 1 and ends on Day 2

					// For midnight-spanning breaks:
					// - break_start must be after clock_in (on Day 1)
					// - break_end must be before clock_out (on Day 2)
					if !breakStartAfterClockIn {
						errs = append(errs, validator.ValidationError{
							Field:   "break_start_time",
							Message: "for breaks spanning midnight, break_start_time must be after clock_in_time",
						})
					}
					if !breakEndBeforeClockOut {
						errs = append(errs, validator.ValidationError{
							Field:   "break_end_time",
							Message: "for breaks spanning midnight, break_end_time must be before clock_out_time",
						})
					}
				}

			} else {
				// ===== SAME-DAY SHIFT VALIDATION =====
				// Example: Clock In 09:00 → Clock Out 17:00
				// Standard sequential validation: clock_in < break_start < break_end < clock_out

				if *r.BreakStartTime < r.ClockInTime {
					errs = append(errs, validator.ValidationError{
						Field:   "break_start_time",
						Message: "break_start_time must be after clock_in_time",
					})
				}

				if *r.BreakStartTime >= *r.BreakEndTime {
					errs = append(errs, validator.ValidationError{
						Field:   "break_start_time",
						Message: "break_start_time must be before break_end_time",
					})
				}

				if *r.BreakEndTime > r.ClockOutTime {
					errs = append(errs, validator.ValidationError{
						Field:   "break_end_time",
						Message: "break_end_time must be before clock_out_time",
					})
				}
			}
		}
	}

	// Validate LocationType - required
	if validator.IsEmpty(r.LocationType) {
		errs = append(errs, validator.ValidationError{
			Field:   "location_type",
			Message: "location_type is required",
		})
	} else if !validator.IsInSlice(r.LocationType, WorkArrangementValues) {
		errs = append(errs, validator.ValidationError{
			Field:   "location_type",
			Message: "location_type must be one of: " + strings.Join(WorkArrangementValues, ", "),
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type WorkScheduleTimeResponse struct {
	ID                string  `json:"id"`
	WorkScheduleID    string  `json:"work_schedule_id"`
	DayOfWeek         int     `json:"day_of_week"`
	DayName           string  `json:"day_name"`
	ClockInTime       string  `json:"clock_in_time"`              // ISO 8601 format
	BreakStartTime    *string `json:"break_start_time,omitempty"` // ISO 8601 format, optional
	BreakEndTime      *string `json:"break_end_time,omitempty"`   // ISO 8601 format, optional
	ClockOutTime      string  `json:"clock_out_time"`             // ISO 8601 format
	IsNextDayCheckout bool    `json:"is_next_day_checkout"`       // New field
	LocationType      string  `json:"location_type"`
	CreatedAt         string  `json:"created_at"` // ISO 8601 format
	UpdatedAt         string  `json:"updated_at"` // ISO 8601 format
}

type CreateWorkScheduleLocationRequest struct {
	WorkScheduleID string  `json:"work_schedule_id"`
	LocationName   string  `json:"location_name"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	RadiusMeters   int     `json:"radius_meters"`
}

func (r *CreateWorkScheduleLocationRequest) Validate() error {
	var errs validator.ValidationErrors

	if validator.IsEmpty(r.WorkScheduleID) {
		errs = append(errs, validator.ValidationError{
			Field:   "work_schedule_id",
			Message: "work_schedule_id is required",
		})
	}
	if validator.IsEmpty(r.LocationName) {
		errs = append(errs, validator.ValidationError{
			Field:   "location_name",
			Message: "location_name is required",
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
	if r.RadiusMeters <= 0 {
		errs = append(errs, validator.ValidationError{
			Field:   "radius_meters",
			Message: "radius_meters must be greater than 0",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type WorkScheduleLocationResponse struct {
	ID             string  `json:"id"`
	WorkScheduleID string  `json:"work_schedule_id"`
	LocationName   string  `json:"location_name"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	RadiusMeters   int     `json:"radius_meters"`
	CreatedAt      string  `json:"created_at"` // ISO 8601 format
	UpdatedAt      string  `json:"updated_at"` // ISO 8601 format
}

type CreateEmployeeScheduleAssignmentRequest struct {
	EmployeeID     string  `json:"employee_id"`
	WorkScheduleID string  `json:"work_schedule_id"`
	StartDate      string  `json:"start_date"` // YYYY-MM-DD format
	EndDate        *string `json:"end_date"`   // YYYY-MM-DD format, optional
}

func (r *CreateEmployeeScheduleAssignmentRequest) Validate() error {
	var errs validator.ValidationErrors

	if validator.IsEmpty(r.EmployeeID) {
		errs = append(errs, validator.ValidationError{
			Field:   "employee_id",
			Message: "employee_id is required",
		})
	}
	if validator.IsEmpty(r.WorkScheduleID) {
		errs = append(errs, validator.ValidationError{
			Field:   "work_schedule_id",
			Message: "work_schedule_id is required",
		})
	}
	if validator.IsEmpty(r.StartDate) {
		errs = append(errs, validator.ValidationError{
			Field:   "start_date",
			Message: "start_date is required",
		})
	} else if _, valid := validator.IsValidDateTime(r.StartDate); !valid {
		errs = append(errs, validator.ValidationError{
			Field:   "start_date",
			Message: "start_date must be a valid date in YYYY-MM-DD format",
		})
	}
	if r.EndDate != nil {
		if _, valid := validator.IsValidDateTime(*r.EndDate); !valid {
			errs = append(errs, validator.ValidationError{
				Field:   "end_date",
				Message: "end_date must be a valid date in YYYY-MM-DD format",
			})
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type EmployeeScheduleAssignmentResponse struct {
	ID             string `json:"id"`
	EmployeeID     string `json:"employee_id"`
	WorkScheduleID string `json:"work_schedule_id"`
	StartDate      string `json:"start_date"` // ISO 8601 format
	EndDate        string `json:"end_date"`   // ISO 8601 format, optional
	CreatedAt      string `json:"created_at"` // ISO 8601 format
	UpdatedAt      string `json:"updated_at"` // ISO 8601 format
}

type UpdateWorkScheduleRequest struct {
	ID                 string  `json:"id"`
	CompanyID          string  `json:"-"`
	Name               *string `json:"name,omitempty"`
	Type               *string `json:"type,omitempty"`
	GracePeriodMinutes *int    `json:"grace_period_minutes,omitempty"`
}

func (r *UpdateWorkScheduleRequest) Validate() error {
	var errs validator.ValidationErrors

	if validator.IsEmpty(r.ID) {
		errs = append(errs, validator.ValidationError{
			Field:   "id",
			Message: "id is required",
		})
	}
	if r.Name != nil && validator.IsEmpty(*r.Name) {
		errs = append(errs, validator.ValidationError{
			Field:   "name",
			Message: "name must not be empty",
		})
	}
	if r.Type != nil {
		if validator.IsEmpty(*r.Type) {
			errs = append(errs, validator.ValidationError{
				Field:   "type",
				Message: "type must not be empty",
			})
		} else if !validator.IsInSlice(*r.Type, WorkArrangementValues) {
			errs = append(errs, validator.ValidationError{
				Field:   "type",
				Message: "type must be one of: " + strings.Join(WorkArrangementValues, ", "),
			})
		}
	}
	if r.GracePeriodMinutes != nil && *r.GracePeriodMinutes < 0 {
		errs = append(errs, validator.ValidationError{
			Field:   "grace_period_minutes",
			Message: "grace_period_minutes must be a non-negative number",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type UpdateWorkScheduleTimeRequest struct {
	ID                string  `json:"id"`
	CompanyID         string  `json:"-"`
	DayOfWeek         *int    `json:"day_of_week,omitempty"`
	ClockInTime       string  `json:"clock_in_time"`              // HH:MM format, required
	ClockOutTime      string  `json:"clock_out_time"`             // HH:MM format, required
	IsNextDayCheckout *bool   `json:"is_next_day_checkout"`       // Required
	BreakStartTime    *string `json:"break_start_time,omitempty"` // HH:MM format, optional
	BreakEndTime      *string `json:"break_end_time,omitempty"`   // HH:MM format, optional
	LocationType      string  `json:"location_type"`              // Required
}

func (r *UpdateWorkScheduleTimeRequest) Validate() error {
	var errs validator.ValidationErrors

	if validator.IsEmpty(r.ID) {
		errs = append(errs, validator.ValidationError{
			Field:   "id",
			Message: "id is required",
		})
	}
	if r.DayOfWeek != nil && (*r.DayOfWeek < 1 || *r.DayOfWeek > 7) {
		errs = append(errs, validator.ValidationError{
			Field:   "day_of_week",
			Message: "day_of_week must be between 1 (Monday) and 7 (Sunday)",
		})
	}

	// Validate ClockInTime - required
	if validator.IsEmpty(r.ClockInTime) {
		errs = append(errs, validator.ValidationError{
			Field:   "clock_in_time",
			Message: "clock_in_time is required",
		})
	} else if _, valid := validator.IsValidTime(r.ClockInTime); !valid {
		errs = append(errs, validator.ValidationError{
			Field:   "clock_in_time",
			Message: "clock_in_time must be a valid time in HH:MM format",
		})
	}

	// Validate ClockOutTime - required
	if validator.IsEmpty(r.ClockOutTime) {
		errs = append(errs, validator.ValidationError{
			Field:   "clock_out_time",
			Message: "clock_out_time is required",
		})
	} else if _, valid := validator.IsValidTime(r.ClockOutTime); !valid {
		errs = append(errs, validator.ValidationError{
			Field:   "clock_out_time",
			Message: "clock_out_time must be a valid time in HH:MM format",
		})
	}

	if r.IsNextDayCheckout == nil {
		errs = append(errs, validator.ValidationError{
			Field:   "is_next_day_checkout",
			Message: "is_next_day_checkout is required",
		})
	}

	// Validate ClockInTime and ClockOutTime based on is_next_day_checkout
	if r.IsNextDayCheckout != nil && !*r.IsNextDayCheckout && r.ClockInTime > r.ClockOutTime {
		errs = append(errs, validator.ValidationError{
			Field:   "clock_in_time",
			Message: "clock_out_time must not be before clock_in_time when is_next_day_checkout is false",
		})
	}
	if r.IsNextDayCheckout != nil && *r.IsNextDayCheckout && r.ClockInTime < r.ClockOutTime {
		errs = append(errs, validator.ValidationError{
			Field:   "clock_in_time",
			Message: "clock_out_time must be before clock_in_time when is_next_day_checkout is true",
		})
	}

	// Validate break times (both must be provided or neither)
	if (r.BreakStartTime != nil && r.BreakEndTime == nil) || (r.BreakStartTime == nil && r.BreakEndTime != nil) {
		errs = append(errs, validator.ValidationError{
			Field:   "break_times",
			Message: "both break_start_time and break_end_time must be provided or neither",
		})
	} else if r.BreakStartTime != nil && r.BreakEndTime != nil {
		_, validStart := validator.IsValidTime(*r.BreakStartTime)
		_, validEnd := validator.IsValidTime(*r.BreakEndTime)

		if !validStart {
			errs = append(errs, validator.ValidationError{
				Field:   "break_start_time",
				Message: "break_start_time must be a valid time in HH:MM format",
			})
		}
		if !validEnd {
			errs = append(errs, validator.ValidationError{
				Field:   "break_end_time",
				Message: "break_end_time must be a valid time in HH:MM format",
			})
		}

		if validStart && validEnd {
			// Validate break times based on shift type
			if r.IsNextDayCheckout != nil && *r.IsNextDayCheckout {
				// ===== OVERNIGHT SHIFT VALIDATION =====
				breakStartAfterClockIn := *r.BreakStartTime >= r.ClockInTime
				breakEndBeforeClockOut := *r.BreakEndTime <= r.ClockOutTime

				validBreakWindow := breakStartAfterClockIn || breakEndBeforeClockOut

				if !validBreakWindow {
					errs = append(errs, validator.ValidationError{
						Field:   "break_times",
						Message: "break must be within shift window (after clock_in or before clock_out)",
					})
				}

				if *r.BreakStartTime < *r.BreakEndTime {
					if breakStartAfterClockIn && *r.BreakEndTime > "23:59" {
						errs = append(errs, validator.ValidationError{
							Field:   "break_end_time",
							Message: "break_end_time cannot extend past midnight if break_start_time is before midnight",
						})
					}
				} else {
					if !breakStartAfterClockIn {
						errs = append(errs, validator.ValidationError{
							Field:   "break_start_time",
							Message: "for breaks spanning midnight, break_start_time must be after clock_in_time",
						})
					}
					if !breakEndBeforeClockOut {
						errs = append(errs, validator.ValidationError{
							Field:   "break_end_time",
							Message: "for breaks spanning midnight, break_end_time must be before clock_out_time",
						})
					}
				}

			} else {
				// ===== SAME-DAY SHIFT VALIDATION =====
				if *r.BreakStartTime < r.ClockInTime {
					errs = append(errs, validator.ValidationError{
						Field:   "break_start_time",
						Message: "break_start_time must be after clock_in_time",
					})
				}

				if *r.BreakStartTime >= *r.BreakEndTime {
					errs = append(errs, validator.ValidationError{
						Field:   "break_start_time",
						Message: "break_start_time must be before break_end_time",
					})
				}

				if *r.BreakEndTime > r.ClockOutTime {
					errs = append(errs, validator.ValidationError{
						Field:   "break_end_time",
						Message: "break_end_time must be before clock_out_time",
					})
				}
			}
		}
	}

	// Validate LocationType - required
	if validator.IsEmpty(r.LocationType) {
		errs = append(errs, validator.ValidationError{
			Field:   "location_type",
			Message: "location_type is required",
		})
	} else if !validator.IsInSlice(r.LocationType, WorkArrangementValues) {
		errs = append(errs, validator.ValidationError{
			Field:   "location_type",
			Message: "location_type must be one of: " + strings.Join(WorkArrangementValues, ", "),
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type UpdateWorkScheduleLocationRequest struct {
	ID           string   `json:"id"`
	CompanyID    string   `json:"-"`
	LocationName *string  `json:"location_name,omitempty"`
	Latitude     *float64 `json:"latitude,omitempty"`
	Longitude    *float64 `json:"longitude,omitempty"`
	RadiusMeters *int     `json:"radius_meters,omitempty"`
}

func (r *UpdateWorkScheduleLocationRequest) Validate() error {
	var errs validator.ValidationErrors

	if validator.IsEmpty(r.ID) {
		errs = append(errs, validator.ValidationError{
			Field:   "id",
			Message: "id is required",
		})
	}
	if r.LocationName != nil && validator.IsEmpty(*r.LocationName) {
		errs = append(errs, validator.ValidationError{
			Field:   "location_name",
			Message: "location_name must not be empty",
		})
	}
	if r.Latitude != nil && (*r.Latitude < -90 || *r.Latitude > 90) {
		errs = append(errs, validator.ValidationError{
			Field:   "latitude",
			Message: "latitude must be between -90 and 90",
		})
	}
	if r.Longitude != nil && (*r.Longitude < -180 || *r.Longitude > 180) {
		errs = append(errs, validator.ValidationError{
			Field:   "longitude",
			Message: "longitude must be between -180 and 180",
		})
	}
	if r.RadiusMeters != nil && *r.RadiusMeters <= 0 {
		errs = append(errs, validator.ValidationError{
			Field:   "radius_meters",
			Message: "radius_meters must be greater than 0",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type UpdateEmployeeScheduleAssignmentRequest struct {
	ID             string `json:"id"`
	EmployeeID     string `json:"employee_id,omitempty"`
	WorkScheduleID string `json:"work_schedule_id,omitempty"`
	StartDate      string `json:"start_date"`
	EndDate        string `json:"end_date"`
}

func (r *UpdateEmployeeScheduleAssignmentRequest) Validate() error {
	var errs validator.ValidationErrors

	if validator.IsEmpty(r.ID) {
		errs = append(errs, validator.ValidationError{
			Field:   "id",
			Message: "id is required",
		})
	}
	if validator.IsEmpty(r.EmployeeID) {
		errs = append(errs, validator.ValidationError{
			Field:   "employee_id",
			Message: "employee_id must not be empty",
		})
	}
	if validator.IsEmpty(r.WorkScheduleID) {
		errs = append(errs, validator.ValidationError{
			Field:   "work_schedule_id",
			Message: "work_schedule_id must not be empty",
		})
	}
	if validator.IsEmpty(r.StartDate) {
		errs = append(errs, validator.ValidationError{
			Field:   "start_date",
			Message: "start_date is required",
		})
	} else if _, valid := validator.IsValidDateTime(r.StartDate); !valid {
		errs = append(errs, validator.ValidationError{
			Field:   "start_date",
			Message: "start_date must be a valid date in YYYY-MM-DD format",
		})
	}
	if validator.IsEmpty(r.EndDate) {
		errs = append(errs, validator.ValidationError{
			Field:   "end_date",
			Message: "end_date is required",
		})
	} else if _, valid := validator.IsValidDateTime(r.EndDate); !valid {
		errs = append(errs, validator.ValidationError{
			Field:   "end_date",
			Message: "end_date must be a valid date in YYYY-MM-DD format",
		})
	} else if r.EndDate < r.StartDate {
		errs = append(errs, validator.ValidationError{
			Field:   "end_date",
			Message: "end_date must not be before start_date",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type AssignScheduleRequest struct {
	WorkScheduleID string  `json:"work_schedule_id"`
	EmployeeID     string  `json:"-"`
	StartDate      string  `json:"start_date"`
	EndDate        *string `json:"end_date,omitempty"`
}

func (r *AssignScheduleRequest) Validate() error {
	var errs validator.ValidationErrors

	if validator.IsEmpty(r.WorkScheduleID) {
		errs = append(errs, validator.ValidationError{
			Field:   "work_schedule_id",
			Message: "work_schedule_id is required",
		})
	}
	if validator.IsEmpty(r.EmployeeID) {
		errs = append(errs, validator.ValidationError{
			Field:   "employee_id",
			Message: "employee_id is required",
		})
	}
	if validator.IsEmpty(r.StartDate) {
		errs = append(errs, validator.ValidationError{
			Field:   "start_date",
			Message: "start_date is required",
		})
	} else if _, valid := validator.IsValidDateTime(r.StartDate); !valid {
		errs = append(errs, validator.ValidationError{
			Field:   "start_date",
			Message: "start_date must be a valid date in YYYY-MM-DD format",
		})
	}
	if r.EndDate != nil {
		if _, valid := validator.IsValidDateTime(*r.EndDate); !valid {
			errs = append(errs, validator.ValidationError{
				Field:   "end_date",
				Message: "end_date must be a valid date in YYYY-MM-DD format",
			})
		} else if *r.EndDate < r.StartDate {
			errs = append(errs, validator.ValidationError{
				Field:   "end_date",
				Message: "end_date must not be before start_date",
			})
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type AssignScheduleResponse struct {
	EmployeeScheduleAssignmentID *string `json:"employee_schedule_assignment_id,omitempty"`
	WorkScheduleID               string  `json:"work_schedule_id"`
	EmployeeID                   string  `json:"employee_id"`
	StartDate                    *string `json:"start_date,omitempty"`
	EndDate                      *string `json:"end_date,omitempty"`
	CreatedAt                    *string `json:"created_at,omitempty"`
	UpdatedAt                    *string `json:"updated_at,omitempty"`
}

// ========================================
// EMPLOYEE SCHEDULE TIMELINE DTOs
// ========================================

type EmployeeScheduleTimelineFilter struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

func (f *EmployeeScheduleTimelineFilter) Validate() error {
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
		f.Limit = 10 // Default limit
	}
	if f.Limit > 100 {
		errs = append(errs, validator.ValidationError{
			Field:   "limit",
			Message: "limit must not exceed 100",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type EmployeeScheduleTimelineResponse struct {
	EmployeeID   string                         `json:"employee_id"`
	EmployeeName string                         `json:"employee_name"`
	TotalCount   int64                          `json:"total_count"`
	Page         int                            `json:"page"`
	Limit        int                            `json:"limit"`
	TotalPages   int                            `json:"total_pages"`
	Showing      string                         `json:"showing"`
	Timeline     []EmployeeScheduleTimelineItem `json:"timeline"`
}

type EmployeeScheduleTimelineItem struct {
	ID               *string           `json:"id"`     // assignment_id for override, null for default
	Type             string            `json:"type"`   // "override" or "default"
	Status           string            `json:"status"` // "active", "upcoming", "past", "fallback"
	DateRange        ScheduleDateRange `json:"date_range"`
	ScheduleSnapshot ScheduleSnapshot  `json:"schedule_snapshot"`
	IsActiveToday    bool              `json:"is_active_today"`
	Actions          ScheduleActions   `json:"actions"`
}

type ScheduleDateRange struct {
	Start *string `json:"start"` // YYYY-MM-DD, can be null for default without history
	End   *string `json:"end"`   // YYYY-MM-DD, null for default (ongoing)
}

type ScheduleSnapshot struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Type               string `json:"type"`
	GracePeriodMinutes int    `json:"grace_period_minutes"`
}

type ScheduleActions struct {
	CanEdit    bool `json:"can_edit"`
	CanDelete  bool `json:"can_delete"`
	CanReplace bool `json:"can_replace,omitempty"` // only for default
}

// ScheduleLocation untuk parsing JSON lokasi
type ScheduleLocation struct {
	Name         string  `json:"name"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	RadiusMeters int     `json:"radius_meters"`
}

// Domain Model (Output bersih untuk Service)
type ActiveSchedule struct {
	ScheduleID         string
	ScheduleName       string
	LocationType       string
	GracePeriodMinutes int
	TimeID             string
	ClockIn            time.Time
	ClockOut           time.Time
	IsNextDayCheckout  bool
	Locations          []ScheduleLocation
}
