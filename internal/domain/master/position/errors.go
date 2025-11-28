package position

import "errors"

var (
	ErrPositionNotFound   = errors.New("position not found")
	ErrPositionNameExists = errors.New("position with this name already exists")
	ErrPositionsNotFound  = errors.New("no positions found")
	ErrUnauthorizedAccess = errors.New("unauthorized access to position")
)
