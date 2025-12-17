package report

import "errors"

var (
	ErrInvalidMonth           = errors.New("month must be between 1 and 12")
	ErrInvalidYear            = errors.New("year must be a valid year")
	ErrInvalidDateRange       = errors.New("end date must be after start date")
	ErrNoDataFound            = errors.New("no data found for the specified criteria")
	ErrReportGenerationFailed = errors.New("failed to generate report")
)
