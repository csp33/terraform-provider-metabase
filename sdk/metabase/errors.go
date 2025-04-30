// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package metabase

import (
	"fmt"
	"net/http"
)

// BaseError is the base error type for all domain-specific errors.
type BaseError struct {
	StatusCode int
	Message    string
}

func (e *BaseError) Error() string {
	return fmt.Sprintf("%s (status code: %d)", e.Message, e.StatusCode)
}

// NotFoundError is returned when a resource is not found (404).
type NotFoundError struct {
	BaseError
}

func NewNotFoundError(message string) *NotFoundError {
	return &NotFoundError{
		BaseError: BaseError{
			StatusCode: http.StatusNotFound,
			Message:    message,
		},
	}
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("Resource not found: %s", e.Message)
}

// ConflictError is returned when there's a conflict (409), e.g., resource already exists.
type ConflictError struct {
	BaseError
}

func NewConflictError(message string) *ConflictError {
	return &ConflictError{
		BaseError: BaseError{
			StatusCode: http.StatusConflict,
			Message:    message,
		},
	}
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("Conflict: %s", e.Message)
}

// BadRequestError is returned for bad requests (400).
type BadRequestError struct {
	BaseError
}

func NewBadRequestError(message string) *BadRequestError {
	return &BadRequestError{
		BaseError: BaseError{
			StatusCode: http.StatusBadRequest,
			Message:    message,
		},
	}
}

func (e *BadRequestError) Error() string {
	return fmt.Sprintf("Bad request: %s", e.Message)
}

// UnauthorizedError is returned for unauthorized requests (401).
type UnauthorizedError struct {
	BaseError
}

func NewUnauthorizedError(message string) *UnauthorizedError {
	return &UnauthorizedError{
		BaseError: BaseError{
			StatusCode: http.StatusUnauthorized,
			Message:    message,
		},
	}
}

// ForbiddenError is returned for forbidden requests (403).
type ForbiddenError struct {
	BaseError
}

func NewForbiddenError(message string) *ForbiddenError {
	return &ForbiddenError{
		BaseError: BaseError{
			StatusCode: http.StatusForbidden,
			Message:    message,
		},
	}
}

func (e *ForbiddenError) Error() string {
	return fmt.Sprintf("Access forbidden: %s", e.Message)
}

// CreateErrorFromStatusCode creates a domain-specific error based on the status code.
func CreateErrorFromStatusCode(statusCode int, message string) error {
	switch statusCode {
	case http.StatusNotFound:
		return NewNotFoundError(message)
	case http.StatusConflict:
		return NewConflictError(message)
	case http.StatusBadRequest:
		return NewBadRequestError(message)
	case http.StatusUnauthorized:
		return NewUnauthorizedError(message)
	case http.StatusForbidden:
		return NewForbiddenError(message)
	default:
		return &BaseError{
			StatusCode: statusCode,
			Message:    message,
		}
	}
}
