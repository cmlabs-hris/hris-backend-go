package leave

import (
	"context"
)

// LeaveTypeRepository - interface for leave_types table
type LeaveTypeRepository interface {
	Create(ctx context.Context, leaveType LeaveType) (LeaveType, error)
	GetByID(ctx context.Context, id string) (LeaveType, error)
	GetByCompanyID(ctx context.Context, companyID string) ([]LeaveType, error)
	Update(ctx context.Context, leaveType LeaveType) error
	Delete(ctx context.Context, id string) error
}

// LeaveQuotaRepository - interface for leave_quotas table
type LeaveQuotaRepository interface {
	Create(ctx context.Context, quota LeaveQuota) (LeaveQuota, error)
	GetByEmployeeAndYear(ctx context.Context, employeeID string, year int) ([]LeaveQuota, error)
	GetByEmployeeTypeYear(ctx context.Context, employeeID, leaveTypeID string, year int) (LeaveQuota, error)
	Update(ctx context.Context, quota LeaveQuota) error
	DecrementQuota(ctx context.Context, quotaID string, days int) error
}

// LeaveRequestRepository - interface for leave_requests table
type LeaveRequestRepository interface {
	Create(ctx context.Context, request LeaveRequest) (LeaveRequest, error)
	GetByID(ctx context.Context, id string) (LeaveRequest, error)
	GetByEmployeeID(ctx context.Context, employeeID string, filter string) ([]LeaveRequest, int64, error)
	GetByCompanyID(ctx context.Context, companyID string, filter string) ([]LeaveRequest, int64, error)
	Update(ctx context.Context, request LeaveRequest) error
	UpdateStatus(ctx context.Context, id, status string, approvedBy *string) error
}
