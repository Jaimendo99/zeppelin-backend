package domain

import "errors"

// Define specific error variables (sentinel errors)
var (
	ErrAuthTokenInvalid      = errors.New("authentication token is invalid or missing")
	ErrAuthTokenMissing      = errors.New("authentication token is missing")
	ErrAuthorizationFailed   = errors.New("user is not authorized for this action")
	ErrRoleExtractionFailed  = errors.New("failed to extract role from claims")
	ErrResourceNotFound      = errors.New("resource not found")              // Example
	ErrValidationFailed      = errors.New("input validation failed")         // Example
	ErrRequiredParamsMissing = errors.New("required parameters are missing") // Example
)
