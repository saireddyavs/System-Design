package repositories

import "errors"

var (
	ErrNotFound          = errors.New("resource not found")
	ErrAlreadyExists      = errors.New("resource already exists")
	ErrInsufficientStock  = errors.New("insufficient stock")
)
