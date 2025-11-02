package postgresql

import (
	"context"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/employee"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
)

type employeeRepositoryImpl struct {
	db *database.DB
}

func NewEmployeeRepository(db *database.DB) employee.EmployeeRepository {
	return &employeeRepositoryImpl{db: db}
}

// GetActiveByCompanyID implements employee.EmployeeRepository.
func (e *employeeRepositoryImpl) GetActiveByCompanyID(ctx context.Context, companyID string) ([]employee.Employee, error) {
	q := GetQuerier(ctx, e.db)

	query := `
		SELECT id, user_id, company_id, work_schedule_id, position_id, grade_id, branch_id, employee_code,
			full_name, nik, gender, phone_number, address, place_of_birth, dob, avatar_url, education,
			hire_date, resignation_date, employment_type, employment_status, warning_letter,
			bank_name, bank_account_holder_name, bank_account_number, created_at, updated_at, deleted_at
		FROM employees
		WHERE company_id = $1 AND employment_status = $2
	`

	rows, err := q.Query(ctx, query, companyID, employee.EmploymentStatusActive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees []employee.Employee
	for rows.Next() {
		var emp employee.Employee
		err := rows.Scan(
			&emp.ID, &emp.UserID, &emp.CompanyID, &emp.WorkScheduleID, &emp.PositionID,
			&emp.GradeID, &emp.BranchID, &emp.EmployeeCode, &emp.FullName, &emp.NIK,
			&emp.Gender, &emp.PhoneNumber, &emp.Address, &emp.PlaceOfBirth, &emp.DOB,
			&emp.AvatarURL, &emp.Education, &emp.HireDate, &emp.ResignationDate,
			&emp.EmploymentType, &emp.EmploymentStatus, &emp.WarningLetter,
			&emp.BankName, &emp.BankAccountHolderName, &emp.BankAccountNumber,
			&emp.CreatedAt, &emp.UpdatedAt, &emp.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		employees = append(employees, emp)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return employees, nil
}

// Create implements employee.EmployeeRepository.
func (e *employeeRepositoryImpl) Create(ctx context.Context, newEmployee employee.Employee) (employee.Employee, error) {
	q := GetQuerier(ctx, e.db)

	query := `
		INSERT INTO employees (
			user_id, company_id, work_schedule_id, position_id, grade_id, branch_id, employee_code,
			full_name, nik, gender, phone_number, address, place_of_birth, dob, avatar_url, education,
			hire_date, resignation_date, employment_type, employment_status, warning_letter,
			bank_name, bank_account_holder_name, bank_account_number
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12, $13, $14, $15, $16, $17,
			$18, $19, $20, $21, $22,
			$23, $24, $25
		)
		RETURNING id, user_id, company_id, work_schedule_id, position_id, grade_id, branch_id, employee_code,
			full_name, nik, gender, phone_number, address, place_of_birth, dob, avatar_url, education,
			hire_date, resignation_date, employment_type, employment_status, warning_letter,
			bank_name, bank_account_holder_name, bank_account_number, created_at, updated_at, deleted_at
	`

	var created employee.Employee
	err := q.QueryRow(ctx, query,
		newEmployee.UserID, newEmployee.CompanyID, newEmployee.WorkScheduleID, newEmployee.PositionID,
		newEmployee.GradeID, newEmployee.BranchID, newEmployee.EmployeeCode, newEmployee.FullName, newEmployee.NIK,
		newEmployee.Gender, newEmployee.PhoneNumber, newEmployee.Address, newEmployee.PlaceOfBirth, newEmployee.DOB,
		newEmployee.AvatarURL, newEmployee.Education, newEmployee.HireDate, newEmployee.ResignationDate,
		newEmployee.EmploymentType, newEmployee.EmploymentStatus, newEmployee.WarningLetter,
		newEmployee.BankName, newEmployee.BankAccountHolderName, newEmployee.BankAccountNumber,
	).Scan(
		&created.ID, &created.UserID, &created.CompanyID, &created.WorkScheduleID, &created.PositionID,
		&created.GradeID, &created.BranchID, &created.EmployeeCode, &created.FullName, &created.NIK,
		&created.Gender, &created.PhoneNumber, &created.Address, &created.PlaceOfBirth, &created.DOB,
		&created.AvatarURL, &created.Education, &created.HireDate, &created.ResignationDate,
		&created.EmploymentType, &created.EmploymentStatus, &created.WarningLetter,
		&created.BankName, &created.BankAccountHolderName, &created.BankAccountNumber,
		&created.CreatedAt, &created.UpdatedAt, &created.DeletedAt,
	)
	if err != nil {
		return employee.Employee{}, err
	}
	return created, nil
}

// ExistsByIDOrCodeOrNIK implements employee.EmployeeRepository.
func (e *employeeRepositoryImpl) ExistsByIDOrCodeOrNIK(ctx context.Context, companyID string, id, employeeCode, nik *string) (bool, error) {
	q := GetQuerier(ctx, e.db)

	var query string
	var arg interface{}

	switch {
	case id != nil:
		query = `SELECT EXISTS(SELECT 1 FROM employees WHERE id = $1 AND company_id = $2)`
		arg = *id
	case employeeCode != nil:
		query = `SELECT EXISTS(SELECT 1 FROM employees WHERE employee_code = $1 AND company_id = $2)`
		arg = *employeeCode
	case nik != nil:
		query = `SELECT EXISTS(SELECT 1 FROM employees WHERE nik = $1 AND company_id = $2)`
		arg = *nik
	default:
		return false, nil
	}

	var exists bool
	err := q.QueryRow(ctx, query, arg, companyID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// GetByEmployeeCode implements employee.EmployeeRepository.
func (e *employeeRepositoryImpl) GetByEmployeeCode(ctx context.Context, companyID, employeeCode string) (employee.Employee, error) {
	q := GetQuerier(ctx, e.db)

	query := `
		SELECT id, user_id, company_id, work_schedule_id, position_id, grade_id, branch_id, employee_code,
			full_name, nik, gender, phone_number, address, place_of_birth, dob, avatar_url, education,
			hire_date, resignation_date, employment_type, employment_status, warning_letter,
			bank_name, bank_account_holder_name, bank_account_number, created_at, updated_at, deleted_at
		FROM employees
		WHERE employee_code = $1 AND company_id = $2
	`

	var found employee.Employee
	err := q.QueryRow(ctx, query, employeeCode, companyID).
		Scan(
			&found.ID, &found.UserID, &found.CompanyID, &found.WorkScheduleID, &found.PositionID,
			&found.GradeID, &found.BranchID, &found.EmployeeCode, &found.FullName, &found.NIK,
			&found.Gender, &found.PhoneNumber, &found.Address, &found.PlaceOfBirth, &found.DOB,
			&found.AvatarURL, &found.Education, &found.HireDate, &found.ResignationDate,
			&found.EmploymentType, &found.EmploymentStatus, &found.WarningLetter,
			&found.BankName, &found.BankAccountHolderName, &found.BankAccountNumber,
			&found.CreatedAt, &found.UpdatedAt, &found.DeletedAt,
		)
	if err != nil {
		return employee.Employee{}, err
	}

	return found, nil
}

// GetByID implements employee.EmployeeRepository.
func (e *employeeRepositoryImpl) GetByID(ctx context.Context, id string) (employee.Employee, error) {
	q := GetQuerier(ctx, e.db)

	query := `
		SELECT id, user_id, company_id, work_schedule_id, position_id, grade_id, branch_id, employee_code,
			full_name, nik, gender, phone_number, address, place_of_birth, dob, avatar_url, education,
			hire_date, resignation_date, employment_type, employment_status, warning_letter,
			bank_name, bank_account_holder_name, bank_account_number, created_at, updated_at, deleted_at
		FROM employees
		WHERE id = $1
	`

	var found employee.Employee
	err := q.QueryRow(ctx, query, id).
		Scan(
			&found.ID, &found.UserID, &found.CompanyID, &found.WorkScheduleID, &found.PositionID,
			&found.GradeID, &found.BranchID, &found.EmployeeCode, &found.FullName, &found.NIK,
			&found.Gender, &found.PhoneNumber, &found.Address, &found.PlaceOfBirth, &found.DOB,
			&found.AvatarURL, &found.Education, &found.HireDate, &found.ResignationDate,
			&found.EmploymentType, &found.EmploymentStatus, &found.WarningLetter,
			&found.BankName, &found.BankAccountHolderName, &found.BankAccountNumber,
			&found.CreatedAt, &found.UpdatedAt, &found.DeletedAt,
		)
	if err != nil {
		return employee.Employee{}, err
	}

	return found, nil
}

// GetByUserID implements employee.EmployeeRepository.
func (e *employeeRepositoryImpl) GetByUserID(ctx context.Context, userID string) (employee.Employee, error) {
	q := GetQuerier(ctx, e.db)

	query := `
		SELECT id, user_id, company_id, work_schedule_id, position_id, grade_id, branch_id, employee_code,
			full_name, nik, gender, phone_number, address, place_of_birth, dob, avatar_url, education,
			hire_date, resignation_date, employment_type, employment_status, warning_letter,
			bank_name, bank_account_holder_name, bank_account_number, created_at, updated_at, deleted_at
		FROM employees
		WHERE user_id = $1
	`

	var found employee.Employee
	err := q.QueryRow(ctx, query, userID).
		Scan(
			&found.ID, &found.UserID, &found.CompanyID, &found.WorkScheduleID, &found.PositionID,
			&found.GradeID, &found.BranchID, &found.EmployeeCode, &found.FullName, &found.NIK,
			&found.Gender, &found.PhoneNumber, &found.Address, &found.PlaceOfBirth, &found.DOB,
			&found.AvatarURL, &found.Education, &found.HireDate, &found.ResignationDate,
			&found.EmploymentType, &found.EmploymentStatus, &found.WarningLetter,
			&found.BankName, &found.BankAccountHolderName, &found.BankAccountNumber,
			&found.CreatedAt, &found.UpdatedAt, &found.DeletedAt,
		)
	if err != nil {
		return employee.Employee{}, err
	}

	return found, nil
}

// Update implements employee.EmployeeRepository.
func (e *employeeRepositoryImpl) Update(ctx context.Context, id string, req employee.UpdateEmployeeRequest) error {
	// TODO IMPLEMENT LATER ON BASED ON WHAT SHOULD BE EDITABLE DATA
	// q := GetQuerier(ctx, e.db)

	// updates := make(map[string]interface{})

	// // Example: Add fields as needed based on UpdateEmployeeRequest
	// if req.FullName != nil {
	// 	if *req.FullName == "" {
	// 		updates["full_name"] = nil
	// 	} else {
	// 		updates["full_name"] = *req.FullName
	// 	}
	// }
	// if req.PhoneNumber != nil {
	// 	if *req.PhoneNumber == "" {
	// 		updates["phone_number"] = nil
	// 	} else {
	// 		updates["phone_number"] = *req.PhoneNumber
	// 	}
	// }
	// Add more fields as needed...

	// if len(updates) == 0 {
	// 	return fmt.Errorf("no updatable fields provided for employee update")
	// }
	// updates["updated_at"] = time.Now()

	// setClauses := make([]string, 0, len(updates))
	// args := make([]interface{}, 0, len(updates)+1)
	// i := 1
	// for col, val := range updates {
	// 	setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, i))
	// 	args = append(args, val)
	// 	i++
	// }

	// sql := "UPDATE employees SET " + strings.Join(setClauses, ", ") + fmt.Sprintf(" WHERE id = $%d", i)
	// args = append(args, id)

	// fmt.Println(sql)
	// var updatedID string
	// if err := q.QueryRow(ctx, sql+" RETURNING id", args...).Scan(&updatedID); err != nil {
	// 	return fmt.Errorf("failed to update employee with id %s: %w", id, err)
	// }
	return nil
}
