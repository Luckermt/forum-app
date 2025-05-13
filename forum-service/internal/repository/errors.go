package repository

import "errors"

var (
	ErrTopicNotFound  = errors.New("topic not found")
	ErrUserNotFound   = errors.New("user not found")
	ErrAccessDenied   = errors.New("access denied")
	ErrAlreadyExists  = errors.New("already exists")
	ErrInvalidRequest = errors.New("invalid request")
)
