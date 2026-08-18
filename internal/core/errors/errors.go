package core_errors

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrRequiredField   = errors.New("field required")
	ErrConflict        = errors.New("conflict")
)
