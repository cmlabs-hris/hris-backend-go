package leave

import (
	"context"
	"time"
)

type LeaveTypeRepository interface {
	Create(ctx context.Context, leaveType LeaveType) (LeaveType, error)
	GetByID(ctx context.Context, id string) (LeaveType, error)
	GetByName(ctx context.Context, companyID, name string) (LeaveType, error)
	GetByCode(ctx context.Context, companyID, code string) (LeaveType, error)
	GetByCompanyID(ctx context.Context, companyID string) ([]LeaveType, error)
	GetActiveByCompanyID(ctx context.Context, companyID string) ([]LeaveType, error)
	Update(ctx context.Context, leaveType UpdateLeaveTypeRequest) error
	Delete(ctx context.Context, id string) error
}

type LeaveQuotaRepository interface {
	Create(ctx context.Context, quota LeaveQuota) (LeaveQuota, error)
	GetByID(ctx context.Context, id string) (LeaveQuota, error)
	GetByEmployeeTypeYear(ctx context.Context, employeeID, leaveTypeID string, year int) (LeaveQuota, error)
	GetByEmployeeYear(ctx context.Context, employeeID string, year int) ([]LeaveQuota, error)
	GetByEmployee(ctx context.Context, employeeID string) ([]LeaveQuota, error)
	GetByCompanyID(ctx context.Context, companyID string) ([]LeaveQuota, error)
	GetByCompanyIDAndYear(ctx context.Context, companyID string, year int) ([]LeaveQuota, error)
	Update(ctx context.Context, quota UpdateLeaveQuotaRequest) error
	AddPendingQuota(ctx context.Context, quotaID string, amount float64) error
	MovePendingToUsed(ctx context.Context, quotaID string, amount float64) error
	RemovePendingQuota(ctx context.Context, quotaID string, amount float64) error
	Delete(ctx context.Context, id string) error
}

type LeaveRequestRepository interface {
	Create(ctx context.Context, req LeaveRequest) (LeaveRequest, error)
	GetByID(ctx context.Context, id string) (LeaveRequest, error)
	GetByEmployeeID(ctx context.Context, employeeID string, filter LeaveRequestFilter) ([]LeaveRequest, int64, error)
	GetByCompanyID(ctx context.Context, companyID string, filter LeaveRequestFilter) ([]LeaveRequest, int64, error)
	GetMyRequests(ctx context.Context, employeeID string, companyID string, filter MyLeaveRequestFilter) ([]LeaveRequest, int64, error)
	Update(ctx context.Context, request UpdateLeaveRequestRequest) error
	CheckOverlapping(ctx context.Context, employeeID string, startDate, endDate time.Time) (bool, error)
	GetMyRequest(ctx context.Context, userID string, companyID string) ([]LeaveRequest, int64, error)
}
