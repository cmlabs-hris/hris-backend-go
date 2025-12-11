package payroll

import "context"

type PayrollService interface {
	// Settings
	GetSettings(ctx context.Context) (PayrollSettingsResponse, error)
	UpdateSettings(ctx context.Context, req UpdatePayrollSettingsRequest) (PayrollSettingsResponse, error)

	// Components
	CreateComponent(ctx context.Context, req CreatePayrollComponentRequest) (PayrollComponentResponse, error)
	GetComponent(ctx context.Context, id string) (PayrollComponentResponse, error)
	ListComponents(ctx context.Context, activeOnly bool) ([]PayrollComponentResponse, error)
	UpdateComponent(ctx context.Context, req UpdatePayrollComponentRequest) error
	DeleteComponent(ctx context.Context, id string) error

	// Employee Components
	AssignComponentToEmployee(ctx context.Context, req AssignComponentRequest) (EmployeeComponentResponse, error)
	GetEmployeeComponents(ctx context.Context, employeeID string) ([]EmployeeComponentResponse, error)
	UpdateEmployeeComponent(ctx context.Context, req UpdateEmployeeComponentRequest) error
	RemoveEmployeeComponent(ctx context.Context, id string) error

	// Payroll Generation & Management
	GeneratePayroll(ctx context.Context, req GeneratePayrollRequest) ([]PayrollRecordResponse, error)
	GetPayrollRecord(ctx context.Context, id string) (PayrollRecordResponse, error)
	ListPayrollRecords(ctx context.Context, filter PayrollFilter) (ListPayrollRecordResponse, error)
	UpdatePayrollRecord(ctx context.Context, req UpdatePayrollRecordRequest) (PayrollRecordResponse, error)
	FinalizePayroll(ctx context.Context, req FinalizePayrollRequest) error
	DeletePayrollRecord(ctx context.Context, id string) error

	// Summary
	GetPayrollSummary(ctx context.Context, month, year int) (PayrollSummaryResponse, error)
}
