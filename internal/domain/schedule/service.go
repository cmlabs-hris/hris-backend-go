package schedule

import (
	"context"
	"time"
)

type ScheduleService interface {
	// Work Schedule
	CreateWorkSchedule(ctx context.Context, req CreateWorkScheduleRequest) (WorkScheduleResponse, error)
	GetWorkSchedule(ctx context.Context, id string) (WorkScheduleResponse, error)
	ListWorkSchedules(ctx context.Context, filter WorkScheduleFilter) (ListWorkScheduleResponse, error)
	UpdateWorkSchedule(ctx context.Context, req UpdateWorkScheduleRequest) error
	DeleteWorkSchedule(ctx context.Context, id string) error

	// Work Schedule Time
	CreateWorkScheduleTime(ctx context.Context, req CreateWorkScheduleTimeRequest) (WorkScheduleTimeResponse, error)
	GetWorkScheduleTime(ctx context.Context, id string) (WorkScheduleTimeResponse, error)
	UpdateWorkScheduleTime(ctx context.Context, req UpdateWorkScheduleTimeRequest) error
	DeleteWorkScheduleTime(ctx context.Context, id string) error

	// Work Schedule Location
	CreateWorkScheduleLocation(ctx context.Context, req CreateWorkScheduleLocationRequest) (WorkScheduleLocationResponse, error)
	GetWorkScheduleLocation(ctx context.Context, id string) (WorkScheduleLocationResponse, error)
	UpdateWorkScheduleLocation(ctx context.Context, req UpdateWorkScheduleLocationRequest) error
	DeleteWorkScheduleLocation(ctx context.Context, id string) error

	// Employee Schedule Assignment
	CreateEmployeeScheduleAssignment(ctx context.Context, req CreateEmployeeScheduleAssignmentRequest) (EmployeeScheduleAssignmentResponse, error)
	GetEmployeeScheduleAssignment(ctx context.Context, id string) (EmployeeScheduleAssignmentResponse, error)
	ListEmployeeScheduleAssignments(ctx context.Context, employeeID string) ([]EmployeeScheduleAssignmentResponse, error)
	UpdateEmployeeScheduleAssignment(ctx context.Context, req UpdateEmployeeScheduleAssignmentRequest) error
	DeleteEmployeeScheduleAssignment(ctx context.Context, id string) error
	GetActiveScheduleForEmployee(ctx context.Context, employeeID string, date time.Time) (WorkScheduleResponse, error)

	// Assign Schedule
	AssignSchedule(ctx context.Context, req AssignScheduleRequest) (AssignScheduleResponse, error)

	// Employee Schedule Timeline
	GetEmployeeScheduleTimeline(ctx context.Context, employeeID string, filter EmployeeScheduleTimelineFilter) (EmployeeScheduleTimelineResponse, error)
}
