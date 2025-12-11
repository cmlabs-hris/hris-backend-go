package employee

import (
	"time"

	"github.com/shopspring/decimal"
)

type Employee struct {
	ID                    string
	UserID                *string
	CompanyID             string
	WorkScheduleID        string
	PositionID            string
	GradeID               string
	BranchID              string
	EmployeeCode          string
	FullName              string
	NIK                   string
	Gender                Gender
	PhoneNumber           string
	Address               *string
	PlaceOfBirth          *string
	DOB                   *time.Time
	AvatarURL             *string
	Education             *string
	HireDate              time.Time
	ResignationDate       *time.Time
	EmploymentType        EmploymentType
	EmploymentStatus      EmploymentStatus
	WarningLetter         *WarningLetter
	BankName              string
	BankAccountHolderName *string
	BankAccountNumber     string
	BaseSalary            *decimal.Decimal
	CreatedAt             time.Time
	UpdatedAt             time.Time
	DeletedAt             *time.Time
}

type Gender string

const (
	Male   Gender = "Male"
	Female Gender = "Female"
)

type EmploymentType string

const (
	EmploymentTypePermanent  EmploymentType = "permanent"
	EmploymentTypeProbation  EmploymentType = "probation"
	EmploymentTypeContract   EmploymentType = "contract"
	EmploymentTypeInternship EmploymentType = "internship"
	EmploymentTypeFreelance  EmploymentType = "freelance"
)

type EmploymentStatus string

const (
	EmploymentStatusActive     EmploymentStatus = "active"
	EmploymentStatusResigned   EmploymentStatus = "resigned"
	EmploymentStatusTerminated EmploymentStatus = "terminated"
)

type WarningLetter string

const (
	WarningLetterLight  WarningLetter = "light"
	WarningLetterMedium WarningLetter = "medium"
	WarningLetterHeavy  WarningLetter = "heavy"
)
