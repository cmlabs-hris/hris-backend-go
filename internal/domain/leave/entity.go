package leave

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// LeaveType entity
type LeaveType struct {
	ID          string
	CompanyID   string
	Name        string
	Code        *string
	Description *string
	Color       *string

	// Policy Rules
	IsActive                    *bool
	RequiresApproval            *bool
	RequiresAttachment          *bool
	AttachmentRequiredAfterDays *int

	// Quota Rules
	HasQuota      *bool
	AccrualMethod *string // 'yearly', 'monthly', 'none'

	// Deduction Rules
	DeductionType *string // 'working_days', 'calendar_days'
	AllowHalfDay  *bool

	// Request Rules
	MaxDaysPerRequest *int
	MinNoticeDays     *int
	MaxAdvanceDays    *int
	AllowBackdate     *bool
	BackdateMaxDays   *int

	// Rollover Rules
	AllowRollover       *bool
	MaxRolloverDays     *int
	RolloverExpiryMonth *int

	// Quota Calculation
	QuotaCalculationType string // 'fixed', 'tenure_based', 'position_based', etc
	QuotaRules           QuotaRules

	CreatedAt time.Time
	UpdatedAt time.Time
}

// QuotaRules represents the JSONB quota calculation rules
type QuotaRules struct {
	Type         string      `json:"type"` // 'fixed', 'tenure', 'position', 'grade', 'employment_type', 'combined'
	Rules        []QuotaRule `json:"rules,omitempty"`
	DefaultQuota float64     `json:"default_quota,omitempty"`
	BaseQuota    float64     `json:"base_quota,omitempty"`
}

// QuotaRule represents individual rule with conditions
type QuotaRule struct {
	// For tenure-based
	MinMonths *int `json:"min_months,omitempty"`
	MaxMonths *int `json:"max_months,omitempty"`

	// For position/grade-based
	PositionIDs []string `json:"position_ids,omitempty"`
	GradeIDs    []string `json:"grade_ids,omitempty"`

	// For employment type-based
	EmploymentType string `json:"employment_type,omitempty"`

	// For combined rules
	Conditions *QuotaConditions `json:"conditions,omitempty"`

	// Resulting quota
	Quota float64 `json:"quota"`
}

// QuotaConditions for complex combined rules
type QuotaConditions struct {
	PositionIDs     []string `json:"position_ids,omitempty"`
	GradeIDs        []string `json:"grade_ids,omitempty"`
	EmploymentType  string   `json:"employment_type,omitempty"`
	MinTenureMonths *int     `json:"min_tenure_months,omitempty"`
	MaxTenureMonths *int     `json:"max_tenure_months,omitempty"`
}

// Value implements driver.Valuer for database storage
func (qr QuotaRules) Value() (driver.Value, error) {
	if qr.Type == "" {
		return nil, nil
	}
	return json.Marshal(qr)
}

// Scan implements sql.Scanner for database retrieval
func (qr *QuotaRules) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan QuotaRules: invalid type")
	}

	return json.Unmarshal(bytes, qr)
}

func ConvertMapToQuotaRules(m map[string]interface{}) (QuotaRules, error) {
	var qr QuotaRules
	b, err := json.Marshal(m)
	if err != nil {
		return qr, err
	}
	err = json.Unmarshal(b, &qr)
	return qr, err
}

// LeaveQuota entity
type LeaveQuota struct {
	ID          string
	EmployeeID  string
	LeaveTypeID string
	Year        int

	OpeningBalance  *int
	EarnedQuota     *int
	RolloverQuota   *int
	AdjustmentQuota *int

	UsedQuota      *float64
	PendingQuota   *float64
	AvailableQuota *float64 // Computed field

	RolloverExpiryDate *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

type LeaveRequestStatus string

const (
	LeaveRequestStatusWaitingApproval LeaveRequestStatus = "waiting_approval"
	LeaveRequestStatusApproved        LeaveRequestStatus = "approved"
	LeaveRequestStatusRejected        LeaveRequestStatus = "rejected"
	LeaveRequestStatusCancelled       LeaveRequestStatus = "cancelled"
)

// LeaveDurationEnum maps to leave_duration_enum in DB
type LeaveDurationEnum string

const (
	LeaveDurationFullDay          LeaveDurationEnum = "full_day"
	LeaveDurationHalfDayMorning   LeaveDurationEnum = "half_day_morning"
	LeaveDurationHalfDayAfternoon LeaveDurationEnum = "half_day_afternoon"
)

// LeaveRequest entity
type LeaveRequest struct {
	ID          string
	EmployeeID  string
	LeaveTypeID string

	StartDate time.Time
	EndDate   time.Time

	DurationType LeaveDurationEnum // 'full_day', 'half_day_morning', 'half_day_afternoon'
	TotalDays    float64
	WorkingDays  float64

	Reason         string
	AttachmentURL  *string
	EmergencyLeave bool
	IsBackdate     bool

	Status          LeaveRequestStatus // 'waiting_approval', 'approved', 'rejected', 'cancelled'
	ApprovedBy      *string
	ApprovedAt      *time.Time
	RejectionReason *string

	CancelledBy        *string
	CancelledAt        *time.Time
	CancellationReason *string

	SubmittedAt time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time

	// Relationships (for responses)
	LeaveTypeName *string
	EmployeeName  *string
}
