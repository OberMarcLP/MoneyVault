package models

import "net/http"

type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(code, message string, status int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: status}
}

// Common application errors
var (
	ErrNotFound      = &AppError{Code: "NOT_FOUND", Message: "resource not found", HTTPStatus: http.StatusNotFound}
	ErrUnauthorized  = &AppError{Code: "UNAUTHORIZED", Message: "authentication required", HTTPStatus: http.StatusUnauthorized}
	ErrForbidden     = &AppError{Code: "FORBIDDEN", Message: "access denied", HTTPStatus: http.StatusForbidden}
	ErrConflict      = &AppError{Code: "CONFLICT", Message: "resource already exists", HTTPStatus: http.StatusConflict}
	ErrInternalError = &AppError{Code: "INTERNAL_ERROR", Message: "an unexpected error occurred", HTTPStatus: http.StatusInternalServerError}
)

func ErrValidation(message string) *AppError {
	return &AppError{Code: "VALIDATION_ERROR", Message: message, HTTPStatus: http.StatusBadRequest}
}

func ErrNotFoundMsg(message string) *AppError {
	return &AppError{Code: "NOT_FOUND", Message: message, HTTPStatus: http.StatusNotFound}
}

func ErrUnauthorizedMsg(message string) *AppError {
	return &AppError{Code: "UNAUTHORIZED", Message: message, HTTPStatus: http.StatusUnauthorized}
}

func ErrConflictMsg(message string) *AppError {
	return &AppError{Code: "CONFLICT", Message: message, HTTPStatus: http.StatusConflict}
}
