package employee

import (
	"context"
)

// EmployeeService defines business logic for employee operations
type EmployeeService interface {
	// SearchEmployees searches for employees with autocomplete (companyID from JWT)
	SearchEmployees(ctx context.Context, req SearchEmployeeRequest) ([]SearchEmployeeResponse, error)

	// GetEmployee retrieves a single employee by ID (with role-based access control)
	GetEmployee(ctx context.Context, id string) (EmployeeResponse, error)

	// CreateEmployee creates a new employee (manager+ only)
	CreateEmployee(ctx context.Context, req CreateEmployeeRequest) (EmployeeResponse, error)

	// UpdateEmployee updates an existing employee (manager+ OR same employee)
	UpdateEmployee(ctx context.Context, req UpdateEmployeeRequest) (EmployeeResponse, error)

	// DeleteEmployee soft deletes an employee (manager+ only)
	DeleteEmployee(ctx context.Context, id string) error

	// ListEmployees lists employees with filters (manager+ only)
	ListEmployees(ctx context.Context, filter EmployeeFilter) (ListEmployeeResponse, error)

	// InactivateEmployee sets resignation_date and employment_status to inactive (manager+ only)
	InactivateEmployee(ctx context.Context, req InactivateEmployeeRequest) (EmployeeResponse, error)

	// UploadAvatar uploads avatar for an employee
	UploadAvatar(ctx context.Context, req UploadAvatarRequest) (EmployeeResponse, error)
}
