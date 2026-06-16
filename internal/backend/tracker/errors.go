package tracker

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound   = errors.New("tracker: not found")
	ErrValidation = errors.New("tracker: validation failed")
	ErrConflict   = errors.New("tracker: conflict")
)

type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	if e == nil {
		return ErrNotFound.Error()
	}
	if e.ID == "" {
		return fmt.Sprintf("%s: %s", ErrNotFound, e.Resource)
	}
	return fmt.Sprintf("%s: %s %q", ErrNotFound, e.Resource, e.ID)
}

func (e *NotFoundError) Is(target error) bool {
	return target == ErrNotFound
}

type ValidationError struct {
	Message string
	Fields  map[string]string
}

func (e *ValidationError) Error() string {
	if e == nil || e.Message == "" {
		return ErrValidation.Error()
	}
	return fmt.Sprintf("%s: %s", ErrValidation, e.Message)
}

func (e *ValidationError) Is(target error) bool {
	return target == ErrValidation
}

type ConflictError struct {
	Resource string
	Field    string
	Value    string
	Message  string
}

func (e *ConflictError) Error() string {
	if e == nil {
		return ErrConflict.Error()
	}
	if e.Message != "" {
		return fmt.Sprintf("%s: %s", ErrConflict, e.Message)
	}
	if e.Resource != "" && e.Field != "" {
		return fmt.Sprintf("%s: %s with %s %q already exists", ErrConflict, e.Resource, e.Field, e.Value)
	}
	return ErrConflict.Error()
}

func (e *ConflictError) Is(target error) bool {
	return target == ErrConflict
}

func validationFailed(fields map[string]string) error {
	return &ValidationError{
		Message: "invalid tracker input",
		Fields:  fields,
	}
}

func notFound(resource string, id string) error {
	return &NotFoundError{Resource: resource, ID: id}
}

func conflict(resource string, field string, value string) error {
	return &ConflictError{Resource: resource, Field: field, Value: value}
}
