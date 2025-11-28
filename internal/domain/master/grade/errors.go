package grade

import "errors"

var (
	ErrGradeNotFound      = errors.New("grade not found")
	ErrGradeNameExists    = errors.New("grade with this name already exists")
	ErrGradesNotFound     = errors.New("no grades found")
	ErrUnauthorizedAccess = errors.New("unauthorized access to grade")
)
