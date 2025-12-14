package postgresql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/employee"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/jackc/pgx/v5"
)

type employeeRepositoryImpl struct {
	db *database.DB
}

// UpdateSchedule implements employee.EmployeeRepository.
func (e *employeeRepositoryImpl) UpdateSchedule(ctx context.Context, id string, workScheduleID string, companyID string) error {
	q := GetQuerier(ctx, e.db)

	query := `
		UPDATE employees
		SET work_schedule_id = $1, updated_at = NOW()
		WHERE id = $2 AND company_id = $3
		RETURNING id
	`

	var updatedID string
	err := q.QueryRow(ctx, query, workScheduleID, id, companyID).Scan(&updatedID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("employee with id %s not found or does not belong to company %s: %w", id, companyID, err)
		}
		return fmt.Errorf("failed to update work schedule for employee with id %s: %w", id, err)
	}

	if len(updatedID) == 0 {
		return fmt.Errorf("update failed, no rows affected")
	}

	return nil
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
			bank_name, bank_account_holder_name, bank_account_number, base_salary, created_at, updated_at, deleted_at
		FROM employees
		WHERE company_id = $1 AND employment_status = $2 AND deleted_at IS NULL
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
			&emp.BaseSalary, &emp.CreatedAt, &emp.UpdatedAt, &emp.DeletedAt,
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
			bank_name, bank_account_holder_name, bank_account_number, base_salary
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21,
			$22, $23, $24, $25
		)
		RETURNING id, user_id, company_id, work_schedule_id, position_id, grade_id, branch_id, employee_code,
			full_name, nik, gender, phone_number, address, place_of_birth, dob, avatar_url, education,
			hire_date, resignation_date, employment_type, employment_status, warning_letter,
			bank_name, bank_account_holder_name, bank_account_number, base_salary, created_at, updated_at, deleted_at
	`

	var created employee.Employee
	err := q.QueryRow(ctx, query,
		newEmployee.UserID, newEmployee.CompanyID, newEmployee.WorkScheduleID, newEmployee.PositionID,
		newEmployee.GradeID, newEmployee.BranchID, newEmployee.EmployeeCode, newEmployee.FullName, newEmployee.NIK,
		newEmployee.Gender, newEmployee.PhoneNumber, newEmployee.Address, newEmployee.PlaceOfBirth, newEmployee.DOB,
		newEmployee.AvatarURL, newEmployee.Education, newEmployee.HireDate, newEmployee.ResignationDate,
		newEmployee.EmploymentType, newEmployee.EmploymentStatus, newEmployee.WarningLetter,
		newEmployee.BankName, newEmployee.BankAccountHolderName, newEmployee.BankAccountNumber, newEmployee.BaseSalary,
	).Scan(
		&created.ID, &created.UserID, &created.CompanyID, &created.WorkScheduleID, &created.PositionID,
		&created.GradeID, &created.BranchID, &created.EmployeeCode, &created.FullName, &created.NIK,
		&created.Gender, &created.PhoneNumber, &created.Address, &created.PlaceOfBirth, &created.DOB,
		&created.AvatarURL, &created.Education, &created.HireDate, &created.ResignationDate,
		&created.EmploymentType, &created.EmploymentStatus, &created.WarningLetter,
		&created.BankName, &created.BankAccountHolderName, &created.BankAccountNumber,
		&created.BaseSalary, &created.CreatedAt, &created.UpdatedAt, &created.DeletedAt,
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
		query = `SELECT EXISTS(SELECT 1 FROM employees WHERE id = $1 AND company_id = $2 AND deleted_at IS NULL)`
		arg = *id
	case employeeCode != nil:
		query = `SELECT EXISTS(SELECT 1 FROM employees WHERE employee_code = $1 AND company_id = $2 AND deleted_at IS NULL)`
		arg = *employeeCode
	case nik != nil:
		query = `SELECT EXISTS(SELECT 1 FROM employees WHERE nik = $1 AND company_id = $2 AND deleted_at IS NULL)`
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
			bank_name, bank_account_holder_name, bank_account_number, base_salary, created_at, updated_at, deleted_at
		FROM employees
		WHERE employee_code = $1 AND company_id = $2 AND deleted_at IS NULL
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
			&found.BaseSalary, &found.CreatedAt, &found.UpdatedAt, &found.DeletedAt,
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
			bank_name, bank_account_holder_name, bank_account_number, base_salary, created_at, updated_at, deleted_at
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
			&found.BaseSalary, &found.CreatedAt, &found.UpdatedAt, &found.DeletedAt,
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
			bank_name, bank_account_holder_name, bank_account_number, base_salary, created_at, updated_at, deleted_at
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
			&found.BaseSalary, &found.CreatedAt, &found.UpdatedAt, &found.DeletedAt,
		)
	if err != nil {
		return employee.Employee{}, err
	}

	return found, nil
}

// Update implements employee.EmployeeRepository.
func (e *employeeRepositoryImpl) Update(ctx context.Context, id string, companyID string, req employee.UpdateEmployeeRequest) error {
	q := GetQuerier(ctx, e.db)

	updates := make(map[string]interface{})

	if req.PositionID != nil {
		if *req.PositionID == "" {
			updates["position_id"] = nil
		} else {
			updates["position_id"] = *req.PositionID
		}
	}
	if req.GradeID != nil {
		if *req.GradeID == "" {
			updates["grade_id"] = nil
		} else {
			updates["grade_id"] = *req.GradeID
		}
	}
	if req.BranchID != nil {
		if *req.BranchID == "" {
			updates["branch_id"] = nil
		} else {
			updates["branch_id"] = *req.BranchID
		}
	}
	if req.EmployeeCode != nil && *req.EmployeeCode != "" {
		updates["employee_code"] = *req.EmployeeCode
	}
	if req.FullName != nil && *req.FullName != "" {
		updates["full_name"] = *req.FullName
	}
	if req.NIK != nil {
		if *req.NIK == "" {
			updates["nik"] = nil
		} else {
			updates["nik"] = *req.NIK
		}
	}
	if req.Gender != nil && *req.Gender != "" {
		updates["gender"] = *req.Gender
	}
	if req.PhoneNumber != nil && *req.PhoneNumber != "" {
		updates["phone_number"] = *req.PhoneNumber
	}
	if req.Address != nil {
		if *req.Address == "" {
			updates["address"] = nil
		} else {
			updates["address"] = *req.Address
		}
	}
	if req.PlaceOfBirth != nil {
		if *req.PlaceOfBirth == "" {
			updates["place_of_birth"] = nil
		} else {
			updates["place_of_birth"] = *req.PlaceOfBirth
		}
	}
	if req.DOB != nil {
		if *req.DOB == "" {
			updates["dob"] = nil
		} else {
			parsedDOB, _ := time.Parse("2006-01-02", *req.DOB)
			updates["dob"] = parsedDOB
		}
	}
	if req.Education != nil {
		if *req.Education == "" {
			updates["education"] = nil
		} else {
			updates["education"] = *req.Education
		}
	}
	if req.HireDate != nil && *req.HireDate != "" {
		parsedHireDate, _ := time.Parse("2006-01-02", *req.HireDate)
		updates["hire_date"] = parsedHireDate
	}
	if req.ResignationDate != nil {
		if *req.ResignationDate == "" {
			updates["resignation_date"] = nil
		} else {
			parsedResignationDate, _ := time.Parse("2006-01-02", *req.ResignationDate)
			updates["resignation_date"] = parsedResignationDate
		}
	}
	if req.EmploymentType != nil && *req.EmploymentType != "" {
		updates["employment_type"] = *req.EmploymentType
	}
	if req.EmploymentStatus != nil && *req.EmploymentStatus != "" {
		updates["employment_status"] = *req.EmploymentStatus
	}
	if req.WarningLetter != nil {
		if *req.WarningLetter == "" {
			updates["warning_letter"] = nil
		} else {
			updates["warning_letter"] = *req.WarningLetter
		}
	}
	if req.BankName != nil {
		if *req.BankName == "" {
			updates["bank_name"] = nil
		} else {
			updates["bank_name"] = *req.BankName
		}
	}
	if req.BankAccountHolderName != nil {
		if *req.BankAccountHolderName == "" {
			updates["bank_account_holder_name"] = nil
		} else {
			updates["bank_account_holder_name"] = *req.BankAccountHolderName
		}
	}
	if req.BankAccountNumber != nil {
		if *req.BankAccountNumber == "" {
			updates["bank_account_number"] = nil
		} else {
			updates["bank_account_number"] = *req.BankAccountNumber
		}
	}
	if req.BaseSalary != nil {
		updates["base_salary"] = *req.BaseSalary
	}

	if len(updates) == 0 {
		return nil // No updates provided
	}
	updates["updated_at"] = time.Now()

	setClauses := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+2)
	i := 1
	for col, val := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}

	sql := fmt.Sprintf("UPDATE employees SET %s WHERE id = $%d AND company_id = $%d AND deleted_at IS NULL RETURNING id", strings.Join(setClauses, ", "), i, i+1)
	args = append(args, id, companyID)

	var updatedID string
	if err := q.QueryRow(ctx, sql, args...).Scan(&updatedID); err != nil {
		if err == pgx.ErrNoRows {
			return employee.ErrEmployeeNotFound
		}
		return fmt.Errorf("failed to update employee with id %s: %w", id, err)
	}
	return nil
}

// GetByIDWithDetails implements employee.EmployeeRepository.
func (e *employeeRepositoryImpl) GetByIDWithDetails(ctx context.Context, id string, companyID string) (employee.EmployeeWithDetails, error) {
	q := GetQuerier(ctx, e.db)

	query := `
		SELECT 
			e.id, e.user_id, e.company_id, e.work_schedule_id, e.position_id, e.grade_id, e.branch_id, 
			e.employee_code, e.full_name, e.nik, e.gender, e.phone_number, e.address, e.place_of_birth, 
			e.dob, e.avatar_url, e.education, e.hire_date, e.resignation_date, e.employment_type, 
			e.employment_status, e.warning_letter, e.bank_name, e.bank_account_holder_name, 
			e.bank_account_number, e.base_salary, e.created_at, e.updated_at, e.deleted_at,
			ws.name AS work_schedule_name,
			p.name AS position_name,
			g.name AS grade_name,
			b.name AS branch_name,
			u.email
		FROM employees e
		LEFT JOIN work_schedules ws ON e.work_schedule_id = ws.id
		LEFT JOIN positions p ON e.position_id = p.id
		LEFT JOIN grades g ON e.grade_id = g.id
		LEFT JOIN branches b ON e.branch_id = b.id
		LEFT JOIN users u ON e.user_id = u.id
		WHERE e.id = $1 AND e.company_id = $2 AND e.deleted_at IS NULL
	`

	var emp employee.EmployeeWithDetails
	err := q.QueryRow(ctx, query, id, companyID).Scan(
		&emp.ID, &emp.UserID, &emp.CompanyID, &emp.WorkScheduleID, &emp.PositionID,
		&emp.GradeID, &emp.BranchID, &emp.EmployeeCode, &emp.FullName, &emp.NIK,
		&emp.Gender, &emp.PhoneNumber, &emp.Address, &emp.PlaceOfBirth, &emp.DOB,
		&emp.AvatarURL, &emp.Education, &emp.HireDate, &emp.ResignationDate,
		&emp.EmploymentType, &emp.EmploymentStatus, &emp.WarningLetter,
		&emp.BankName, &emp.BankAccountHolderName, &emp.BankAccountNumber,
		&emp.BaseSalary, &emp.CreatedAt, &emp.UpdatedAt, &emp.DeletedAt,
		&emp.WorkScheduleName, &emp.PositionName, &emp.GradeName, &emp.BranchName,
		&emp.Email,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return employee.EmployeeWithDetails{}, employee.ErrEmployeeNotFound
		}
		return employee.EmployeeWithDetails{}, fmt.Errorf("failed to get employee: %w", err)
	}

	return emp, nil
}

// Search implements employee.EmployeeRepository.
func (e *employeeRepositoryImpl) Search(ctx context.Context, queryStr string, companyID string, limit int) ([]employee.EmployeeWithDetails, error) {
	q := GetQuerier(ctx, e.db)

	query := `
		SELECT 
			e.id, e.user_id, e.company_id, e.work_schedule_id, e.position_id, e.grade_id, e.branch_id, 
			e.employee_code, e.full_name, e.nik, e.gender, e.phone_number, e.address, e.place_of_birth, 
			e.dob, e.avatar_url, e.education, e.hire_date, e.resignation_date, e.employment_type, 
			e.employment_status, e.warning_letter, e.bank_name, e.bank_account_holder_name, 
			e.bank_account_number, e.base_salary, e.created_at, e.updated_at, e.deleted_at,
			ws.name AS work_schedule_name,
			p.name AS position_name,
			g.name AS grade_name,
			b.name AS branch_name
		FROM employees e
		LEFT JOIN work_schedules ws ON e.work_schedule_id = ws.id
		LEFT JOIN positions p ON e.position_id = p.id
		LEFT JOIN grades g ON e.grade_id = g.id
		LEFT JOIN branches b ON e.branch_id = b.id
		WHERE e.company_id = $1 
			AND e.deleted_at IS NULL
			AND (
				e.full_name ILIKE $2 
				OR e.employee_code ILIKE $2
			)
		ORDER BY e.full_name ASC
		LIMIT $3
	`

	searchPattern := "%" + queryStr + "%"
	rows, err := q.Query(ctx, query, companyID, searchPattern, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search employees: %w", err)
	}
	defer rows.Close()

	var employees []employee.EmployeeWithDetails
	for rows.Next() {
		var emp employee.EmployeeWithDetails
		err := rows.Scan(
			&emp.ID, &emp.UserID, &emp.CompanyID, &emp.WorkScheduleID, &emp.PositionID,
			&emp.GradeID, &emp.BranchID, &emp.EmployeeCode, &emp.FullName, &emp.NIK,
			&emp.Gender, &emp.PhoneNumber, &emp.Address, &emp.PlaceOfBirth, &emp.DOB,
			&emp.AvatarURL, &emp.Education, &emp.HireDate, &emp.ResignationDate,
			&emp.EmploymentType, &emp.EmploymentStatus, &emp.WarningLetter,
			&emp.BankName, &emp.BankAccountHolderName, &emp.BankAccountNumber,
			&emp.BaseSalary, &emp.CreatedAt, &emp.UpdatedAt, &emp.DeletedAt,
			&emp.WorkScheduleName, &emp.PositionName, &emp.GradeName, &emp.BranchName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan employee: %w", err)
		}
		employees = append(employees, emp)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return employees, nil
}

// List implements employee.EmployeeRepository.
func (e *employeeRepositoryImpl) List(ctx context.Context, filter employee.EmployeeFilter, companyID string) ([]employee.EmployeeWithDetails, int64, error) {
	q := GetQuerier(ctx, e.db)

	// Build WHERE conditions
	conditions := []string{"e.company_id = $1", "e.deleted_at IS NULL"}
	args := []interface{}{companyID}
	argIdx := 2

	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(e.full_name ILIKE $%d OR e.employee_code ILIKE $%d OR e.nik ILIKE $%d)", argIdx, argIdx, argIdx))
		args = append(args, "%"+*filter.Search+"%")
		argIdx++
	}
	if filter.WorkScheduleID != nil && *filter.WorkScheduleID != "" {
		conditions = append(conditions, fmt.Sprintf("e.work_schedule_id = $%d", argIdx))
		args = append(args, *filter.WorkScheduleID)
		argIdx++
	}
	if filter.PositionID != nil && *filter.PositionID != "" {
		conditions = append(conditions, fmt.Sprintf("e.position_id = $%d", argIdx))
		args = append(args, *filter.PositionID)
		argIdx++
	}
	if filter.GradeID != nil && *filter.GradeID != "" {
		conditions = append(conditions, fmt.Sprintf("e.grade_id = $%d", argIdx))
		args = append(args, *filter.GradeID)
		argIdx++
	}
	if filter.BranchID != nil && *filter.BranchID != "" {
		conditions = append(conditions, fmt.Sprintf("e.branch_id = $%d", argIdx))
		args = append(args, *filter.BranchID)
		argIdx++
	}
	if filter.EmploymentType != nil && *filter.EmploymentType != "" {
		conditions = append(conditions, fmt.Sprintf("e.employment_type = $%d", argIdx))
		args = append(args, *filter.EmploymentType)
		argIdx++
	}
	if filter.EmploymentStatus != nil && *filter.EmploymentStatus != "" {
		conditions = append(conditions, fmt.Sprintf("e.employment_status = $%d", argIdx))
		args = append(args, *filter.EmploymentStatus)
		argIdx++
	}
	if filter.WarningLetter != nil && *filter.WarningLetter != "" {
		conditions = append(conditions, fmt.Sprintf("e.warning_letter = $%d", argIdx))
		args = append(args, *filter.WarningLetter)
		argIdx++
	}
	if filter.HireDateFrom != nil && *filter.HireDateFrom != "" {
		conditions = append(conditions, fmt.Sprintf("e.hire_date >= $%d", argIdx))
		args = append(args, *filter.HireDateFrom)
		argIdx++
	}
	if filter.HireDateTo != nil && *filter.HireDateTo != "" {
		conditions = append(conditions, fmt.Sprintf("e.hire_date <= $%d", argIdx))
		args = append(args, *filter.HireDateTo)
		argIdx++
	}
	if filter.ResignationDateFrom != nil && *filter.ResignationDateFrom != "" {
		conditions = append(conditions, fmt.Sprintf("e.resignation_date >= $%d", argIdx))
		args = append(args, *filter.ResignationDateFrom)
		argIdx++
	}
	if filter.ResignationDateTo != nil && *filter.ResignationDateTo != "" {
		conditions = append(conditions, fmt.Sprintf("e.resignation_date <= $%d", argIdx))
		args = append(args, *filter.ResignationDateTo)
		argIdx++
	}
	if filter.DOBFrom != nil && *filter.DOBFrom != "" {
		conditions = append(conditions, fmt.Sprintf("e.dob >= $%d", argIdx))
		args = append(args, *filter.DOBFrom)
		argIdx++
	}
	if filter.DOBTo != nil && *filter.DOBTo != "" {
		conditions = append(conditions, fmt.Sprintf("e.dob <= $%d", argIdx))
		args = append(args, *filter.DOBTo)
		argIdx++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM employees e WHERE %s", whereClause)
	var total int64
	err := q.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count employees: %w", err)
	}

	// Validate sort column
	validSortColumns := map[string]string{
		"full_name":         "e.full_name",
		"employee_code":     "e.employee_code",
		"hire_date":         "e.hire_date",
		"created_at":        "e.created_at",
		"employment_status": "e.employment_status",
	}
	sortColumn, ok := validSortColumns[filter.SortBy]
	if !ok {
		sortColumn = "e.created_at"
	}

	sortOrder := "DESC"
	if strings.ToUpper(filter.SortOrder) == "ASC" {
		sortOrder = "ASC"
	}

	// Main query with pagination
	offset := (filter.Page - 1) * filter.Limit
	query := fmt.Sprintf(`
		SELECT 
			e.id, e.user_id, e.company_id, e.work_schedule_id, e.position_id, e.grade_id, e.branch_id, 
			e.employee_code, e.full_name, e.nik, e.gender, e.phone_number, e.address, e.place_of_birth, 
			e.dob, e.avatar_url, e.education, e.hire_date, e.resignation_date, e.employment_type, 
			e.employment_status, e.warning_letter, e.bank_name, e.bank_account_holder_name, 
			e.bank_account_number, e.base_salary, e.created_at, e.updated_at, e.deleted_at,
			ws.name AS work_schedule_name,
			p.name AS position_name,
			g.name AS grade_name,
			b.name AS branch_name
		FROM employees e
		LEFT JOIN work_schedules ws ON e.work_schedule_id = ws.id
		LEFT JOIN positions p ON e.position_id = p.id
		LEFT JOIN grades g ON e.grade_id = g.id
		LEFT JOIN branches b ON e.branch_id = b.id
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortColumn, sortOrder, argIdx, argIdx+1)

	args = append(args, filter.Limit, offset)

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list employees: %w", err)
	}
	defer rows.Close()

	var employees []employee.EmployeeWithDetails
	for rows.Next() {
		var emp employee.EmployeeWithDetails
		err := rows.Scan(
			&emp.ID, &emp.UserID, &emp.CompanyID, &emp.WorkScheduleID, &emp.PositionID,
			&emp.GradeID, &emp.BranchID, &emp.EmployeeCode, &emp.FullName, &emp.NIK,
			&emp.Gender, &emp.PhoneNumber, &emp.Address, &emp.PlaceOfBirth, &emp.DOB,
			&emp.AvatarURL, &emp.Education, &emp.HireDate, &emp.ResignationDate,
			&emp.EmploymentType, &emp.EmploymentStatus, &emp.WarningLetter,
			&emp.BankName, &emp.BankAccountHolderName, &emp.BankAccountNumber,
			&emp.BaseSalary, &emp.CreatedAt, &emp.UpdatedAt, &emp.DeletedAt,
			&emp.WorkScheduleName, &emp.PositionName, &emp.GradeName, &emp.BranchName,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan employee: %w", err)
		}
		employees = append(employees, emp)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return employees, total, nil
}

// SoftDelete implements employee.EmployeeRepository.
func (e *employeeRepositoryImpl) SoftDelete(ctx context.Context, id string, companyID string) error {
	q := GetQuerier(ctx, e.db)

	query := `
		UPDATE employees 
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND company_id = $2 AND deleted_at IS NULL
		RETURNING id
	`

	var deletedID string
	err := q.QueryRow(ctx, query, id, companyID).Scan(&deletedID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return employee.ErrEmployeeNotFound
		}
		return fmt.Errorf("failed to soft delete employee: %w", err)
	}

	return nil
}

// UpdateAvatar implements employee.EmployeeRepository.
func (e *employeeRepositoryImpl) UpdateAvatar(ctx context.Context, id string, companyID string, avatarURL string) error {
	q := GetQuerier(ctx, e.db)

	query := `
		UPDATE employees 
		SET avatar_url = $1, updated_at = NOW()
		WHERE id = $2 AND company_id = $3 AND deleted_at IS NULL
		RETURNING id
	`

	var updatedID string
	err := q.QueryRow(ctx, query, avatarURL, id, companyID).Scan(&updatedID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return employee.ErrEmployeeNotFound
		}
		return fmt.Errorf("failed to update avatar: %w", err)
	}

	return nil
}

// Inactivate implements employee.EmployeeRepository.
func (e *employeeRepositoryImpl) Inactivate(ctx context.Context, id string, companyID string, resignationDate string) error {
	q := GetQuerier(ctx, e.db)

	parsedDate, _ := time.Parse("2006-01-02", resignationDate)

	query := `
		UPDATE employees 
		SET resignation_date = $1, employment_status = $2, updated_at = NOW()
		WHERE id = $3 AND company_id = $4 AND deleted_at IS NULL
		RETURNING id
	`

	var updatedID string
	err := q.QueryRow(ctx, query, parsedDate, employee.EmploymentStatusResigned, id, companyID).Scan(&updatedID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return employee.ErrEmployeeNotFound
		}
		return fmt.Errorf("failed to inactivate employee: %w", err)
	}

	return nil
}

// LinkUser implements employee.EmployeeRepository.
func (e *employeeRepositoryImpl) LinkUser(ctx context.Context, employeeID, userID, companyID string) error {
	q := GetQuerier(ctx, e.db)

	query := `
		UPDATE employees 
		SET user_id = $1, updated_at = NOW()
		WHERE id = $2 AND company_id = $3 AND deleted_at IS NULL
		RETURNING id
	`

	var updatedID string
	err := q.QueryRow(ctx, query, userID, employeeID, companyID).Scan(&updatedID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return employee.ErrEmployeeNotFound
		}
		return fmt.Errorf("failed to link user to employee: %w", err)
	}

	return nil
}

// GetManagersByCompanyID retrieves all managers and owners for a company (for notifications)
func (e *employeeRepositoryImpl) GetManagersByCompanyID(ctx context.Context, companyID string) ([]employee.Employee, error) {
	q := GetQuerier(ctx, e.db)

	query := `
		SELECT e.id, e.user_id, e.company_id, e.work_schedule_id, e.position_id, e.grade_id, e.branch_id, e.employee_code,
			e.full_name, e.nik, e.gender, e.phone_number, e.address, e.place_of_birth, e.dob, e.avatar_url, e.education,
			e.hire_date, e.resignation_date, e.employment_type, e.employment_status, e.warning_letter,
			e.bank_name, e.bank_account_holder_name, e.bank_account_number, e.base_salary, e.created_at, e.updated_at, e.deleted_at
		FROM employees e
		INNER JOIN users u ON e.user_id = u.id
		WHERE e.company_id = $1 
			AND e.employment_status = $2 
			AND e.deleted_at IS NULL
			AND e.user_id IS NOT NULL
			AND u.role IN ('owner', 'manager')
	`

	rows, err := q.Query(ctx, query, companyID, employee.EmploymentStatusActive)
	if err != nil {
		return nil, fmt.Errorf("failed to get managers: %w", err)
	}
	defer rows.Close()

	var managers []employee.Employee
	for rows.Next() {
		var emp employee.Employee
		err := rows.Scan(
			&emp.ID, &emp.UserID, &emp.CompanyID, &emp.WorkScheduleID, &emp.PositionID,
			&emp.GradeID, &emp.BranchID, &emp.EmployeeCode, &emp.FullName, &emp.NIK,
			&emp.Gender, &emp.PhoneNumber, &emp.Address, &emp.PlaceOfBirth, &emp.DOB,
			&emp.AvatarURL, &emp.Education, &emp.HireDate, &emp.ResignationDate,
			&emp.EmploymentType, &emp.EmploymentStatus, &emp.WarningLetter,
			&emp.BankName, &emp.BankAccountHolderName, &emp.BankAccountNumber,
			&emp.BaseSalary, &emp.CreatedAt, &emp.UpdatedAt, &emp.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan manager: %w", err)
		}
		managers = append(managers, emp)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate managers: %w", err)
	}

	return managers, nil
}
