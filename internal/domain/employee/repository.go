package employee

import "context"

// EmployeeWithDetails contains employee data with joined related names
type EmployeeWithDetails struct {
	Employee
	WorkScheduleName *string
	PositionName     *string
	GradeName        *string
	BranchName       *string
}

type EmployeeRepository interface {
	// Basic CRUD
	GetByID(ctx context.Context, id string) (Employee, error)
	GetByUserID(ctx context.Context, userID string) (Employee, error)
	GetByEmployeeCode(ctx context.Context, companyID string, employeeCode string) (Employee, error)
	Create(ctx context.Context, newEmployee Employee) (Employee, error)
	Update(ctx context.Context, id string, companyID string, req UpdateEmployeeRequest) error

	// Extended operations
	ExistsByIDOrCodeOrNIK(ctx context.Context, companyID string, id, employeeCode, nik *string) (bool, error)
	GetActiveByCompanyID(ctx context.Context, companyID string) ([]Employee, error)
	UpdateSchedule(ctx context.Context, id string, workScheduleID string, companyID string) error
	LinkUser(ctx context.Context, employeeID, userID, companyID string) error

	// New operations with company filter and JOINs
	GetByIDWithDetails(ctx context.Context, id string, companyID string) (EmployeeWithDetails, error)
	Search(ctx context.Context, query string, companyID string, limit int) ([]EmployeeWithDetails, error)
	List(ctx context.Context, filter EmployeeFilter, companyID string) ([]EmployeeWithDetails, int64, error)
	SoftDelete(ctx context.Context, id string, companyID string) error
	UpdateAvatar(ctx context.Context, id string, companyID string, avatarURL string) error
	Inactivate(ctx context.Context, id string, companyID string, resignationDate string) error
}
