package payroll

import "context"

// PayrollRepository defines data access methods for payroll.
// All methods include companyID parameter to prevent cross-company data access attacks.
type PayrollRepository interface {
	// Settings
	GetSettings(ctx context.Context, companyID string) (PayrollSettings, error)
	UpsertSettings(ctx context.Context, settings PayrollSettings) (PayrollSettings, error)

	// Components
	CreateComponent(ctx context.Context, component PayrollComponent) (PayrollComponent, error)
	GetComponentByID(ctx context.Context, id string, companyID string) (PayrollComponent, error)
	GetComponentsByCompanyID(ctx context.Context, companyID string, activeOnly bool) ([]PayrollComponent, error)
	UpdateComponent(ctx context.Context, companyID string, req UpdatePayrollComponentRequest) error
	DeleteComponent(ctx context.Context, id string, companyID string) error

	// Employee Components
	AssignComponentToEmployee(ctx context.Context, assignment EmployeePayrollComponent, companyID string) (EmployeePayrollComponent, error)
	GetEmployeeComponents(ctx context.Context, employeeID string, companyID string, activeOnly bool) ([]EmployeePayrollComponent, error)
	GetEmployeeComponentByID(ctx context.Context, id string, companyID string) (EmployeePayrollComponent, error)
	UpdateEmployeeComponent(ctx context.Context, companyID string, req UpdateEmployeeComponentRequest) error
	RemoveEmployeeComponent(ctx context.Context, id string, companyID string) error

	// Payroll Records
	CreatePayrollRecord(ctx context.Context, record PayrollRecord) (PayrollRecord, error)
	GetPayrollRecordByID(ctx context.Context, id string, companyID string) (PayrollRecord, error)
	GetPayrollRecordByEmployeePeriod(ctx context.Context, employeeID string, month, year int, companyID string) (PayrollRecord, error)
	ListPayrollRecords(ctx context.Context, companyID string, filter PayrollFilter) ([]PayrollRecord, int64, error)
	UpdatePayrollRecord(ctx context.Context, companyID string, req UpdatePayrollRecordRequest) error
	FinalizePayrollRecords(ctx context.Context, ids []string, paidBy string, companyID string) error
	DeletePayrollRecord(ctx context.Context, id string, companyID string) error

	// Aggregations
	GetAttendanceSummary(ctx context.Context, companyID string, month, year int, employeeIDs []string) ([]AttendanceSummary, error)
	GetPayrollSummary(ctx context.Context, companyID string, month, year int) (PayrollSummaryResponse, error)
}
