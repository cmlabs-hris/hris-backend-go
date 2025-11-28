package postgresql

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/leave"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/jackc/pgx/v5"
)

type leaveTypeRepositoryImpl struct {
	db *database.DB
}

// GetByName implements leave.LeaveTypeRepository.
func (l *leaveTypeRepositoryImpl) GetByName(ctx context.Context, companyID string, name string) (leave.LeaveType, error) {
	q := GetQuerier(ctx, l.db)
	query := `
		SELECT id, company_id, name, code, description, color,
			   is_active, requires_approval, requires_attachment, attachment_required_after_days,
			   has_quota, accrual_method,
			   deduction_type, allow_half_day,
			   max_days_per_request, min_notice_days, max_advance_days, allow_backdate, backdate_max_days,
			   allow_rollover, max_rollover_days, rollover_expiry_month,
			   quota_calculation_type, quota_rules,
			   created_at, updated_at
		FROM leave_types
		WHERE company_id = $1 AND name = $2
	`
	var lt leave.LeaveType
	var quotaRulesJSON []byte

	err := q.QueryRow(ctx, query, companyID, name).Scan(
		&lt.ID, &lt.CompanyID, &lt.Name, &lt.Code, &lt.Description, &lt.Color,
		&lt.IsActive, &lt.RequiresApproval, &lt.RequiresAttachment, &lt.AttachmentRequiredAfterDays,
		&lt.HasQuota, &lt.AccrualMethod,
		&lt.DeductionType, &lt.AllowHalfDay,
		&lt.MaxDaysPerRequest, &lt.MinNoticeDays, &lt.MaxAdvanceDays, &lt.AllowBackdate, &lt.BackdateMaxDays,
		&lt.AllowRollover, &lt.MaxRolloverDays, &lt.RolloverExpiryMonth,
		&lt.QuotaCalculationType, &quotaRulesJSON,
		&lt.CreatedAt, &lt.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return leave.LeaveType{}, pgx.ErrNoRows
		}
		return leave.LeaveType{}, err
	}

	if quotaRulesJSON != nil {
		json.Unmarshal(quotaRulesJSON, &lt.QuotaRules)
	}

	return lt, nil
}

func NewLeaveTypeRepository(db *database.DB) leave.LeaveTypeRepository {
	return &leaveTypeRepositoryImpl{db: db}
}

// GetByCode implements leave.LeaveTypeRepository.
func (l *leaveTypeRepositoryImpl) GetByCode(ctx context.Context, companyID string, code string) (leave.LeaveType, error) {
	panic("unimplemented")
}

// Delete implements leave.LeaveTypeRepository.
func (l *leaveTypeRepositoryImpl) Delete(ctx context.Context, id string) error {
	q := GetQuerier(ctx, l.db)
	query := `
		DELETE FROM leave_types
		WHERE id = $1
	`
	commandTag, err := q.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return fmt.Errorf("leave type with id %s not found", id)
	}
	return nil
}

// Create implements leave.LeaveTypeRepository.
func (l *leaveTypeRepositoryImpl) Create(ctx context.Context, leaveType leave.LeaveType) (leave.LeaveType, error) {
	q := GetQuerier(ctx, l.db)

	quotaRulesJSON, _ := json.Marshal(leaveType.QuotaRules)

	query := `
		INSERT INTO leave_types (
			id, company_id, name,
			quota_calculation_type, quota_rules,
			created_at, updated_at
		) VALUES (
			uuidv7(), $1, $2, $3, $4,
			NOW(), NOW()
		) RETURNING id, created_at, updated_at
	`

	err := q.QueryRow(ctx, query,
		leaveType.CompanyID, leaveType.Name,
		leaveType.QuotaCalculationType, quotaRulesJSON,
	).Scan(&leaveType.ID, &leaveType.CreatedAt, &leaveType.UpdatedAt)

	if err != nil {
		return leave.LeaveType{}, err
	}

	return leaveType, nil
}

// GetByCompanyID implements leave.LeaveTypeRepository.
func (l *leaveTypeRepositoryImpl) GetByCompanyID(ctx context.Context, companyID string) ([]leave.LeaveType, error) {
	q := GetQuerier(ctx, l.db)
	query := `
		SELECT id, company_id, name, code, description, color,
			   is_active, requires_approval, requires_attachment, attachment_required_after_days,
			   has_quota, accrual_method,
			   deduction_type, allow_half_day,
			   max_days_per_request, min_notice_days, max_advance_days, allow_backdate, backdate_max_days,
			   allow_rollover, max_rollover_days, rollover_expiry_month,
			   quota_calculation_type, quota_rules,
			   created_at, updated_at
		FROM leave_types
		WHERE company_id = $1
		ORDER BY name
	`
	rows, err := q.Query(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var leaveTypes []leave.LeaveType
	for rows.Next() {
		var lt leave.LeaveType
		var quotaRulesJSON []byte

		if err := rows.Scan(
			&lt.ID, &lt.CompanyID, &lt.Name, &lt.Code, &lt.Description, &lt.Color,
			&lt.IsActive, &lt.RequiresApproval, &lt.RequiresAttachment, &lt.AttachmentRequiredAfterDays,
			&lt.HasQuota, &lt.AccrualMethod,
			&lt.DeductionType, &lt.AllowHalfDay,
			&lt.MaxDaysPerRequest, &lt.MinNoticeDays, &lt.MaxAdvanceDays, &lt.AllowBackdate, &lt.BackdateMaxDays,
			&lt.AllowRollover, &lt.MaxRolloverDays, &lt.RolloverExpiryMonth,
			&lt.QuotaCalculationType, &quotaRulesJSON,
			&lt.CreatedAt, &lt.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if quotaRulesJSON != nil {
			json.Unmarshal(quotaRulesJSON, &lt.QuotaRules)
		}

		leaveTypes = append(leaveTypes, lt)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return leaveTypes, nil
}

// GetByID implements leave.LeaveTypeRepository.
func (l *leaveTypeRepositoryImpl) GetByID(ctx context.Context, id string) (leave.LeaveType, error) {
	q := GetQuerier(ctx, l.db)
	query := `
		SELECT id, company_id, name, code, description, color,
			   is_active, requires_approval, requires_attachment, attachment_required_after_days,
			   has_quota, accrual_method,
			   deduction_type, allow_half_day,
			   max_days_per_request, min_notice_days, max_advance_days, allow_backdate, backdate_max_days,
			   allow_rollover, max_rollover_days, rollover_expiry_month,
			   quota_calculation_type, quota_rules,
			   created_at, updated_at
		FROM leave_types
		WHERE id = $1
	`

	var lt leave.LeaveType
	var quotaRulesJSON []byte

	err := q.QueryRow(ctx, query, id).Scan(
		&lt.ID, &lt.CompanyID, &lt.Name, &lt.Code, &lt.Description, &lt.Color,
		&lt.IsActive, &lt.RequiresApproval, &lt.RequiresAttachment, &lt.AttachmentRequiredAfterDays,
		&lt.HasQuota, &lt.AccrualMethod,
		&lt.DeductionType, &lt.AllowHalfDay,
		&lt.MaxDaysPerRequest, &lt.MinNoticeDays, &lt.MaxAdvanceDays, &lt.AllowBackdate, &lt.BackdateMaxDays,
		&lt.AllowRollover, &lt.MaxRolloverDays, &lt.RolloverExpiryMonth,
		&lt.QuotaCalculationType, &quotaRulesJSON,
		&lt.CreatedAt, &lt.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return leave.LeaveType{}, leave.ErrLeaveTypeNotFound
		}
		return leave.LeaveType{}, err
	}

	if quotaRulesJSON != nil {
		json.Unmarshal(quotaRulesJSON, &lt.QuotaRules)
	}

	return lt, nil
}

func (l *leaveTypeRepositoryImpl) GetActiveByCompanyID(ctx context.Context, companyID string) ([]leave.LeaveType, error) {
	q := GetQuerier(ctx, l.db)

	query := `
		SELECT id, company_id, name, code, description, color,
			   is_active, requires_approval, requires_attachment, attachment_required_after_days,
			   has_quota, accrual_method,
			   deduction_type, allow_half_day,
			   max_days_per_request, min_notice_days, max_advance_days, allow_backdate, backdate_max_days,
			   allow_rollover, max_rollover_days, rollover_expiry_month,
			   quota_calculation_type, quota_rules,
			   created_at, updated_at
		FROM leave_types
		WHERE company_id = $1 AND is_active = true
		ORDER BY name
	`

	rows, err := q.Query(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var leaveTypes []leave.LeaveType

	for rows.Next() {
		var lt leave.LeaveType
		var quotaRulesJSON []byte

		err := rows.Scan(
			&lt.ID, &lt.CompanyID, &lt.Name, &lt.Code, &lt.Description, &lt.Color,
			&lt.IsActive, &lt.RequiresApproval, &lt.RequiresAttachment, &lt.AttachmentRequiredAfterDays,
			&lt.HasQuota, &lt.AccrualMethod,
			&lt.DeductionType, &lt.AllowHalfDay,
			&lt.MaxDaysPerRequest, &lt.MinNoticeDays, &lt.MaxAdvanceDays, &lt.AllowBackdate, &lt.BackdateMaxDays,
			&lt.AllowRollover, &lt.MaxRolloverDays, &lt.RolloverExpiryMonth,
			&lt.QuotaCalculationType, &quotaRulesJSON,
			&lt.CreatedAt, &lt.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if quotaRulesJSON != nil {
			json.Unmarshal(quotaRulesJSON, &lt.QuotaRules)
		}

		leaveTypes = append(leaveTypes, lt)
	}

	return leaveTypes, nil
}

// Update implements leave.LeaveTypeRepository.
func (l *leaveTypeRepositoryImpl) Update(ctx context.Context, leaveType leave.UpdateLeaveTypeRequest) error {
	q := GetQuerier(ctx, l.db)

	updates := make([]string, 0)
	args := make([]interface{}, 0)
	argIdx := 1

	// Basic Info
	if leaveType.Name != nil {
		updates = append(updates, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *leaveType.Name)
		argIdx++
	}
	if leaveType.Code != nil {
		updates = append(updates, fmt.Sprintf("code = $%d", argIdx))
		args = append(args, *leaveType.Code)
		argIdx++
	}
	if leaveType.Description != nil {
		updates = append(updates, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *leaveType.Description)
		argIdx++
	}
	if leaveType.Color != nil {
		updates = append(updates, fmt.Sprintf("color = $%d", argIdx))
		args = append(args, *leaveType.Color)
		argIdx++
	}

	// Policy Rules
	if leaveType.IsActive != nil {
		updates = append(updates, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *leaveType.IsActive)
		argIdx++
	}
	if leaveType.RequiresApproval != nil {
		updates = append(updates, fmt.Sprintf("requires_approval = $%d", argIdx))
		args = append(args, *leaveType.RequiresApproval)
		argIdx++
	}
	if leaveType.RequiresAttachment != nil {
		updates = append(updates, fmt.Sprintf("requires_attachment = $%d", argIdx))
		args = append(args, *leaveType.RequiresAttachment)
		argIdx++
	}
	if leaveType.AttachmentRequiredAfterDays != nil {
		updates = append(updates, fmt.Sprintf("attachment_required_after_days = $%d", argIdx))
		args = append(args, *leaveType.AttachmentRequiredAfterDays)
		argIdx++
	}

	// Quota Rules
	if leaveType.HasQuota != nil {
		updates = append(updates, fmt.Sprintf("has_quota = $%d", argIdx))
		args = append(args, *leaveType.HasQuota)
		argIdx++
	}
	if leaveType.AccrualMethod != nil {
		updates = append(updates, fmt.Sprintf("accrual_method = $%d", argIdx))
		args = append(args, *leaveType.AccrualMethod)
		argIdx++
	}

	// Deduction Rules
	if leaveType.DeductionType != nil {
		updates = append(updates, fmt.Sprintf("deduction_type = $%d", argIdx))
		args = append(args, *leaveType.DeductionType)
		argIdx++
	}
	if leaveType.AllowHalfDay != nil {
		updates = append(updates, fmt.Sprintf("allow_half_day = $%d", argIdx))
		args = append(args, *leaveType.AllowHalfDay)
		argIdx++
	}

	// Request Rules
	if leaveType.MaxDaysPerRequest != nil {
		updates = append(updates, fmt.Sprintf("max_days_per_request = $%d", argIdx))
		args = append(args, *leaveType.MaxDaysPerRequest)
		argIdx++
	}
	if leaveType.MinNoticeDays != nil {
		updates = append(updates, fmt.Sprintf("min_notice_days = $%d", argIdx))
		args = append(args, *leaveType.MinNoticeDays)
		argIdx++
	}
	if leaveType.MaxAdvanceDays != nil {
		updates = append(updates, fmt.Sprintf("max_advance_days = $%d", argIdx))
		args = append(args, *leaveType.MaxAdvanceDays)
		argIdx++
	}
	if leaveType.AllowBackdate != nil {
		updates = append(updates, fmt.Sprintf("allow_backdate = $%d", argIdx))
		args = append(args, *leaveType.AllowBackdate)
		argIdx++
	}
	if leaveType.BackdateMaxDays != nil {
		updates = append(updates, fmt.Sprintf("backdate_max_days = $%d", argIdx))
		args = append(args, *leaveType.BackdateMaxDays)
		argIdx++
	}

	// Rollover Rules
	if leaveType.AllowRollover != nil {
		updates = append(updates, fmt.Sprintf("allow_rollover = $%d", argIdx))
		args = append(args, *leaveType.AllowRollover)
		argIdx++
	}
	if leaveType.MaxRolloverDays != nil {
		updates = append(updates, fmt.Sprintf("max_rollover_days = $%d", argIdx))
		args = append(args, *leaveType.MaxRolloverDays)
		argIdx++
	}
	if leaveType.RolloverExpiryMonth != nil {
		updates = append(updates, fmt.Sprintf("rollover_expiry_month = $%d", argIdx))
		args = append(args, *leaveType.RolloverExpiryMonth)
		argIdx++
	}

	// Quota Calculation
	if leaveType.QuotaCalculationType != nil {
		updates = append(updates, fmt.Sprintf("quota_calculation_type = $%d", argIdx))
		args = append(args, *leaveType.QuotaCalculationType)
		argIdx++
	}
	if leaveType.QuotaRules != nil {
		quotaRulesJSON, _ := json.Marshal(leaveType.QuotaRules)
		updates = append(updates, fmt.Sprintf("quota_rules = $%d", argIdx))
		args = append(args, quotaRulesJSON)
		argIdx++
	}

	if len(updates) == 0 {
		return fmt.Errorf("no updatable fields provided for leave type update")
	}

	updates = append(updates, fmt.Sprintf("updated_at = $%d", argIdx))
	args = append(args, time.Now())
	argIdx++

	args = append(args, leaveType.ID)

	sql := "UPDATE leave_types SET " + strings.Join(updates, ", ") + fmt.Sprintf(" WHERE id = $%d RETURNING id", argIdx)

	var updatedID string
	if err := q.QueryRow(ctx, sql, args...).Scan(&updatedID); err != nil {
		return fmt.Errorf("failed to update leave type with id %s: %w", leaveType.ID, err)
	}
	return nil
}
