package entity

import "errors"

var (
	ErrInvalidRequest    = errors.New("invalid request")
	ErrInternalError     = errors.New("internal error")
	ErrResourceExhausted = errors.New("resource exhausted")
	ErrDeadlineExceeded  = errors.New("deadline exceeded")
)
