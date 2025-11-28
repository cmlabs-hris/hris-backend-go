package branch

import "errors"

var (
	ErrBranchNotFound     = errors.New("branch not found")
	ErrBranchNameExists   = errors.New("branch with this name already exists")
	ErrBranchesNotFound   = errors.New("no branches found")
	ErrUnauthorizedAccess = errors.New("unauthorized access to branch")
)
