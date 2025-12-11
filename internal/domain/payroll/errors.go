package payroll

import "errors"

var (
	ErrPayrollSettingsNotFound    = errors.New("payroll settings not found")
	ErrPayrollComponentNotFound   = errors.New("payroll component not found")
	ErrPayrollComponentNameExists = errors.New("payroll component name already exists")
	ErrPayrollRecordNotFound      = errors.New("payroll record not found")
	ErrPayrollRecordAlreadyExists = errors.New("payroll record already exists for this period")
	ErrPayrollRecordAlreadyPaid   = errors.New("payroll record already paid, cannot modify")
	ErrInvalidPeriod              = errors.New("invalid payroll period")
	ErrEmployeeHasNoBaseSalary    = errors.New("employee has no base salary configured")
	ErrCannotDeletePaidRecord     = errors.New("cannot delete paid payroll record")
	ErrEmployeeComponentNotFound  = errors.New("employee component assignment not found")
	ErrEmployeeNotFound           = errors.New("employee not found")
	ErrInvalidComponentType       = errors.New("invalid component type")
)
