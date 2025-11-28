package schedule

import (
	"context"
	"time"
)

type WorkScheduleRepository interface {
	Create(ctx context.Context, workSchedule WorkSchedule) (WorkSchedule, error)
	GetByID(ctx context.Context, id string, companyID string) (WorkSchedule, error)
	GetByCompanyID(ctx context.Context, companyID string, filter WorkScheduleFilter) ([]WorkSchedule, int64, error)
	Update(ctx context.Context, req UpdateWorkScheduleRequest) (WorkSchedule, error)
	Delete(ctx context.Context, id, companyID string) error
	SoftDelete(ctx context.Context, id, companyID string) error
	GetEmployeeScheduleTimeline(ctx context.Context, employeeID, companyID string, page, limit int) ([]EmployeeScheduleTimelineItem, int64, string, error)
	GetActiveSchedule(ctx context.Context, employeeID string, date time.Time, companyID string) (*ActiveSchedule, error)
}

type WorkScheduleTimeRepository interface {
	Create(ctx context.Context, workScheduleTime WorkScheduleTime, companyID string) (WorkScheduleTime, error)
	GetByID(ctx context.Context, id string, companyID string) (WorkScheduleTime, error)
	GetByWorkScheduleID(ctx context.Context, workScheduleID, companyID string) ([]WorkScheduleTime, error)
	GetTimeByScheduleAndDay(ctx context.Context, scheduleID string, dayOfWeek int, companyID string) (WorkScheduleTime, error)
	Update(ctx context.Context, req UpdateWorkScheduleTimeRequest) error
	Delete(ctx context.Context, id, companyID string) error
}

type WorkScheduleLocationRepository interface {
	Create(ctx context.Context, workScheduleLocation WorkScheduleLocation, companyID string) (WorkScheduleLocation, error)
	GetByID(ctx context.Context, id string, companyID string) (WorkScheduleLocation, error)
	GetByWorkScheduleID(ctx context.Context, workScheduleID, companyID string) ([]WorkScheduleLocation, error)
	Update(ctx context.Context, req UpdateWorkScheduleLocationRequest) error
	Delete(ctx context.Context, id, companyID string) error
	BulkDeleteByWorkScheduleID(ctx context.Context, workScheduleID, companyID string) error
}

type EmployeeScheduleAssignmentRepository interface {
	Create(ctx context.Context, assignment EmployeeScheduleAssignment, companyID string) (EmployeeScheduleAssignment, error)
	GetByID(ctx context.Context, id string, companyID string) (EmployeeScheduleAssignment, error)
	GetByEmployeeID(ctx context.Context, employeeID string) ([]EmployeeScheduleAssignment, error)
	GetActiveSchedule(ctx context.Context, employeeID string, date time.Time) (WorkSchedule, error)
	GetScheduleAssignments(ctx context.Context, employeeID string, startDate, endDate time.Time) ([]EmployeeScheduleAssignment, error)
	Update(ctx context.Context, req UpdateEmployeeScheduleAssignmentRequest, companyID string) error
	Delete(ctx context.Context, id string, companyID string) error
	DeleteFutureAssignments(ctx context.Context, startDate time.Time, employeeID, companyID string) error
}
