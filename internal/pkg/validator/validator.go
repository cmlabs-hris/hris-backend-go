package validator

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ValidationError struct {
	Field   string
	Message string
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var msgs []string
	for _, err := range v {
		msgs = append(msgs, err.Field+": "+err.Message)
	}
	return strings.Join(msgs, "; ")
}

func (v ValidationErrors) ToMap() map[string]string {
	result := make(map[string]string)
	for _, err := range v {
		result[err.Field] = err.Message
	}
	return result
}

// IsEmpty checks if a string is empty after trimming whitespace.
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Email validation
func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// UUIDv7 regex: version 7 (the 15th character must be '7'), all lowercase hex digits.
var uuidv7Regex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// UUIDv7 validation
func IsValidUUID(uuid string) bool {
	return uuidv7Regex.MatchString(strings.ToLower(uuid))
}

// Numeric validation
var numericRegex = regexp.MustCompile(`^[0-9]+$`)

func IsNumeric(s string) bool {
	return numericRegex.MatchString(s)
}

// Date validation
func IsValidDate(dateStr string) (time.Time, bool) {
	date, err := time.Parse("2006-01-02", dateStr)
	return date, err == nil
}

// NIK validation (Indonesian ID)
func IsValidNIK(nik string) bool {
	return len(nik) == 16 && IsNumeric(nik)
}

// Phone number validation
func IsValidPhoneNumber(phone string) bool {
	// Remove spaces and dashes
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")

	if len(phone) < 10 || len(phone) > 13 {
		return false
	}

	// Must start with 08, 62, or +62
	if strings.HasPrefix(phone, "08") ||
		strings.HasPrefix(phone, "62") ||
		strings.HasPrefix(phone, "+62") {
		cleanPhone := strings.TrimPrefix(strings.TrimPrefix(phone, "+"), "62")
		return IsNumeric(cleanPhone)
	}

	return false
}

// Slice contains check
func IsInSlice(value string, slice []string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// Username validation: 3-50 chars, A-Z, a-z, 0-9, ., _, -
var companyUsernameRegex = regexp.MustCompile(`^[A-Za-z0-9._-]{3,50}$`)

func IsValidCompanyUsername(companyUsername string) bool {
	return companyUsernameRegex.MatchString(companyUsername)
}

var employeeCodeRegex = regexp.MustCompile(`^\d{4}-\d{4}$`)

func IsValidEmployeeCode(code string) bool {
	return employeeCodeRegex.MatchString(code)
}

type Date time.Time

// ParseDate parses a date string in "YYYY-MM-DD" format and returns a Date type.
func ParseDate(dateStr string) (Date, error) {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return Date{}, err
	}
	return Date(t), nil
}

// Before reports whether the date d is before u.
func (d Date) Before(u Date) bool {
	return time.Time(d).Before(time.Time(u))
}

// Itoa converts an integer to a string.
func Itoa(i int) string {
	return strconv.Itoa(i)
}

// IsValidDateTime checks if a string is a valid date in "YYYY-MM-DD" format.
func IsValidDateTime(dateTimeStr string) (time.Time, bool) {
	t, err := time.Parse("2006-01-02", dateTimeStr)
	if err == nil {
		return t, true
	}
	return time.Time{}, false
}

// IsValidTime checks if a string is a valid time in HH:MM or HH:MM:SS format.
// Accepts formats like: "09:00", "09:00:00", "23:59:59"
func IsValidTime(timeStr string) (time.Time, bool) {
	// Try HH:MM:SS format first
	t, err := time.Parse("15:04:05", timeStr)
	if err == nil {
		return t, true
	}

	// Try HH:MM format (will default seconds to 00)
	t, err = time.Parse("15:04", timeStr)
	if err == nil {
		return t, true
	}

	return time.Time{}, false
}
