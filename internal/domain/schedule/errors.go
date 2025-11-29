package schedule

import "errors"

var (
	// Work Schedule Errors
	ErrWorkScheduleNotFound       = errors.New("work schedule not found")
	ErrWorkScheduleNameExists     = errors.New("work schedule with this name already exists")
	ErrWorkScheduleAlreadyDeleted = errors.New("work schedule not found or already deleted")

	// Work Schedule Time Errors
	ErrWorkScheduleTimeNotFound = errors.New("work schedule time not found")
	ErrWorkScheduleTimeExists   = errors.New("work schedule time already exists")
	ErrInvalidLocationType      = errors.New("invalid location type for work schedule")
	ErrMismatchedLocationType   = errors.New("mismatched location type for work schedule")

	// Work Schedule Location Errors
	ErrWorkScheduleLocationNotFound = errors.New("work schedule location not found")
	ErrInvalidWorkScheduleType      = errors.New("work schedule type must be 'WFO' or match the required location type")

	// Employee Schedule Assignment Errors
	ErrEmployeeScheduleAssignmentNotFound = errors.New("employee schedule assignment not found")
	ErrOverlappingScheduleAssignment      = errors.New("overlapping schedule assignment detected")

	// Employee Schedule Timeline Errors
	ErrEmployeeScheduleTimelineNotFound = errors.New("employee schedule timeline not found")

	// Validation Errors
	ErrEmployeeIDRequired = errors.New("employee ID is required")
	ErrInvalidDateFormat  = errors.New("invalid date format, use YYYY-MM-DD")
	// Request Data Errors
	ErrInvalidRequestData = errors.New("invalid request data")
)
