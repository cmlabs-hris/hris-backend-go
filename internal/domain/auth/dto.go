package auth

type RegisterRequest struct {
	CompanyName     string `json:"company_name"`
	CompanyUsername string `json:"company_username"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginEmployeeCodeRequest struct {
	EmployeeCode string `json:"employee_code"`
	Password     string `json:"password"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type VerifyEmailRequest struct {
	Token string `json:"token"`
}
