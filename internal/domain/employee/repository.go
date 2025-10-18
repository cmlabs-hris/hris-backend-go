package employee

import "context"

type EmployeeRepository interface {
	GetByID(ctx context.Context, id string) (Employee, error)
	GetByEmployeeCode(ctx context.Context, companyID string, employeeCode string) (Employee, error)
	Create(ctx context.Context, newEmployee Employee) (Employee, error)
	ExistsByIDOrCodeOrNIK(ctx context.Context, companyID string, id, employeeCode, nik *string) (bool, error)
	Update(ctx context.Context, id string, req UpdateEmployeeRequest) error
}
