package models

import "errors"

var (
	ErrCountryNotFound = errors.New("country not found")
	ErrInvalidRequest  = errors.New("invalid request")
	ErrAPIFailure      = errors.New("external API failure")
	ErrTimeout         = errors.New("request timeout")
)
