package employee

type UpdateEmployeeRequest struct {
	WorkScheduleID        *string `json:"work_schedule_id,omitempty"`
	PositionID            string  `json:"position_id"`
	GradeID               string  `json:"grade_id,omitempty"`
	BranchID              string  `json:"branch_id,omitempty"`
	EmployeeCode          string  `json:"employee_code"`
	FullName              string  `json:"full_name"`
	NIK                   string  `json:"nik"`
	Gender                string  `json:"gender"`
	PhoneNumber           string  `json:"phone_number"`
	Address               *string `json:"address,omitempty"`
	PlaceOfBirth          *string `json:"place_of_birth,omitempty"`
	DOB                   *string `json:"dob,omitempty"`
	AvatarURL             *string `json:"avatar_url,omitempty"`
	Education             *string `json:"education,omitempty"`
	HireDate              string  `json:"hire_date"`
	ResignationDate       *string `json:"resignation_date,omitempty"`
	EmploymentType        string  `json:"employment_type"`
	EmploymentStatus      string  `json:"employment_status"`
	WarningLetter         *string `json:"warning_letter,omitempty"`
	BankName              string  `json:"bank_name"`
	BankAccountHolderName *string `json:"bank_account_holder_name,omitempty"`
	BankAccountNumber     string  `json:"bank_account_number"`
}

// harusnya semua bisa diedit gak sih kan ini editable
