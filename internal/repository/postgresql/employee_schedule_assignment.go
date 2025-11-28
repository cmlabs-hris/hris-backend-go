package postgresql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/schedule"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/jackc/pgx/v5"
)

type employeeScheduleAssignmentRepository struct {
	db *database.DB
}

// DeleteFutureAssignments implements schedule.EmployeeScheduleAssignmentRepository.
func (e *employeeScheduleAssignmentRepository) DeleteFutureAssignments(ctx context.Context, startDate time.Time, employeeID string, companyID string) error {
	q := GetQuerier(ctx, e.db)

	query := `
		DELETE FROM employee_schedule_assignments
		WHERE employee_id = $1 AND start_date >= $2
		AND EXISTS (
			SELECT 1 FROM employees
			WHERE id = $1 AND company_id = $3
		)
		RETURNING employee_id
	`

	var returnedEmployeeID string
	err := q.QueryRow(ctx, query, employeeID, startDate, companyID).Scan(&returnedEmployeeID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("no future assignments found for employee_id %s starting from %s: %w", employeeID, startDate, err)
		}
		return fmt.Errorf("failed to delete future assignments: %w", err)
	}

	return nil
}

// Create implements schedule.EmployeeScheduleAssignmentRepository.
func (e *employeeScheduleAssignmentRepository) Create(ctx context.Context, assignment schedule.EmployeeScheduleAssignment, companyID string) (schedule.EmployeeScheduleAssignment, error) {
	q := GetQuerier(ctx, e.db)

	query := `
		INSERT INTO employee_schedule_assignments (
			id, employee_id, work_schedule_id, start_date, end_date
		)
		SELECT uuidv7(), $1, $2, $3, $4
		FROM employees
		WHERE id = $1 AND company_id = $5
		RETURNING id
	`

	err := q.QueryRow(ctx, query,
		assignment.EmployeeID, assignment.WorkScheduleID,
		assignment.StartDate, assignment.EndDate, companyID,
	).Scan(&assignment.ID)

	if err != nil {
		return schedule.EmployeeScheduleAssignment{}, err
	}

	return assignment, nil
}

// Delete implements schedule.EmployeeScheduleAssignmentRepository.
func (e *employeeScheduleAssignmentRepository) Delete(ctx context.Context, id string, companyID string) error {
	q := GetQuerier(ctx, e.db)

	query := `
		DELETE FROM employee_schedule_assignments
		WHERE id = $1
		AND EXISTS (
			SELECT 1 FROM employees
			WHERE id = employee_schedule_assignments.employee_id AND company_id = $2
		)
	`

	commandTag, err := q.Exec(ctx, query, id, companyID)
	if err != nil {
		return fmt.Errorf("failed to delete assignment with id %s: %w", id, err)
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// GetActiveSchedule implements schedule.EmployeeScheduleAssignmentRepository.
func (e *employeeScheduleAssignmentRepository) GetActiveSchedule(ctx context.Context, employeeID string, date time.Time) (schedule.WorkSchedule, error) {
	q := GetQuerier(ctx, e.db)

	// Priority 1: Check assignments
	queryAssignment := `
		SELECT ws.id, ws.company_id, ws.name, ws.type, ws.created_at, ws.updated_at
		FROM employee_schedule_assignments esa
		JOIN work_schedules ws ON esa.work_schedule_id = ws.id
		WHERE esa.employee_id = $1
		  AND $2 BETWEEN esa.start_date AND COALESCE(esa.end_date, '9999-12-31'::date)
		ORDER BY esa.start_date DESC
		LIMIT 1
	`

	var ws schedule.WorkSchedule
	err := q.QueryRow(ctx, queryAssignment, employeeID, date).Scan(
		&ws.ID, &ws.CompanyID, &ws.Name, &ws.Type, &ws.CreatedAt, &ws.UpdatedAt,
	)

	if err == nil {
		return ws, nil
	}

	// Priority 2: Fallback to default
	queryDefault := `
		SELECT ws.id, ws.company_id, ws.name, ws.type, ws.created_at, ws.updated_at
		FROM employees e
		JOIN work_schedules ws ON e.work_schedule_id = ws.id
		WHERE e.id = $1
	`

	err = q.QueryRow(ctx, queryDefault, employeeID).Scan(
		&ws.ID, &ws.CompanyID, &ws.Name, &ws.Type, &ws.CreatedAt, &ws.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return schedule.WorkSchedule{}, fmt.Errorf("no schedule found for employee")
		}
		return schedule.WorkSchedule{}, err
	}

	return ws, nil
}

// GetByEmployeeID implements schedule.EmployeeScheduleAssignmentRepository.
func (e *employeeScheduleAssignmentRepository) GetByEmployeeID(ctx context.Context, employeeID string) ([]schedule.EmployeeScheduleAssignment, error) {
	q := GetQuerier(ctx, e.db)

	query := `
		SELECT id, employee_id, work_schedule_id, start_date, end_date, created_at, updated_at
		FROM employee_schedule_assignments
		WHERE employee_id = $1
		ORDER BY start_date DESC
	`

	rows, err := q.Query(ctx, query, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []schedule.EmployeeScheduleAssignment
	for rows.Next() {
		var a schedule.EmployeeScheduleAssignment
		if err := rows.Scan(&a.ID, &a.EmployeeID, &a.WorkScheduleID, &a.StartDate, &a.EndDate, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		assignments = append(assignments, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return assignments, nil
}

// GetByID implements schedule.EmployeeScheduleAssignmentRepository.
func (e *employeeScheduleAssignmentRepository) GetByID(ctx context.Context, id string, companyID string) (schedule.EmployeeScheduleAssignment, error) {
	q := GetQuerier(ctx, e.db)

	query := `
		SELECT esa.id, esa.employee_id, esa.work_schedule_id, esa.start_date, esa.end_date, esa.created_at, esa.updated_at
		FROM employee_schedule_assignments esa
		JOIN employees emp ON esa.employee_id = emp.id
		WHERE esa.id = $1 AND emp.company_id = $2
	`

	var a schedule.EmployeeScheduleAssignment
	err := q.QueryRow(ctx, query, id, companyID).Scan(
		&a.ID, &a.EmployeeID, &a.WorkScheduleID, &a.StartDate, &a.EndDate, &a.CreatedAt, &a.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return schedule.EmployeeScheduleAssignment{}, fmt.Errorf("assignment not found")
		}
		return schedule.EmployeeScheduleAssignment{}, err
	}

	return a, nil
}

// GetScheduleAssignments implements schedule.EmployeeScheduleAssignmentRepository.
func (e *employeeScheduleAssignmentRepository) GetScheduleAssignments(ctx context.Context, employeeID string, startDate time.Time, endDate time.Time) ([]schedule.EmployeeScheduleAssignment, error) {
	q := GetQuerier(ctx, e.db)

	query := `
		SELECT id, employee_id, work_schedule_id, start_date, end_date, created_at, updated_at
		FROM employee_schedule_assignments
		WHERE employee_id = $1
		  AND (
			  (start_date BETWEEN $2 AND $3) OR
			  (end_date BETWEEN $2 AND $3) OR
			  (start_date <= $2 AND COALESCE(end_date, '9999-12-31'::date) >= $3)
		  )
		ORDER BY start_date
	`

	rows, err := q.Query(ctx, query, employeeID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []schedule.EmployeeScheduleAssignment
	for rows.Next() {
		var a schedule.EmployeeScheduleAssignment
		if err := rows.Scan(&a.ID, &a.EmployeeID, &a.WorkScheduleID, &a.StartDate, &a.EndDate, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		assignments = append(assignments, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return assignments, nil
}

func (e *employeeScheduleAssignmentRepository) Update(ctx context.Context, req schedule.UpdateEmployeeScheduleAssignmentRequest, companyID string) error {
	q := GetQuerier(ctx, e.db)

	updates := make([]string, 0)
	args := make([]interface{}, 0)
	argIdx := 1

	updates = append(updates, fmt.Sprintf("employee_id = $%d", argIdx))
	args = append(args, req.EmployeeID)
	argIdx++

	updates = append(updates, fmt.Sprintf("work_schedule_id = $%d", argIdx))
	args = append(args, req.WorkScheduleID)
	argIdx++

	updates = append(updates, fmt.Sprintf("start_date = $%d", argIdx))
	args = append(args, req.StartDate)
	argIdx++

	updates = append(updates, fmt.Sprintf("end_date = $%d", argIdx))
	args = append(args, req.EndDate)
	argIdx++

	updates = append(updates, fmt.Sprintf("updated_at = $%d", argIdx))
	args = append(args, time.Now())
	argIdx++

	args = append(args, req.ID)
	idIdx := argIdx
	argIdx++

	args = append(args, companyID)

	sql := "UPDATE employee_schedule_assignments SET " + strings.Join(updates, ", ") +
		fmt.Sprintf(" WHERE id = $%d AND EXISTS (SELECT 1 FROM employees WHERE id = employee_schedule_assignments.employee_id AND company_id = $%d) RETURNING id", idIdx, argIdx)

	var updatedID string
	if err := q.QueryRow(ctx, sql, args...).Scan(&updatedID); err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("assignment not found or does not belong to the specified company: %w", err)
		}
		return fmt.Errorf("failed to update assignment with id %s: %w", req.ID, err)
	}

	return nil
}

func NewEmployeeScheduleAssignmentRepository(db *database.DB) schedule.EmployeeScheduleAssignmentRepository {
	return &employeeScheduleAssignmentRepository{db: db}
}
