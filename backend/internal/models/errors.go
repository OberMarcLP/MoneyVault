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

func ErrValidation(message string) *AppError {
	return &AppError{Code: "VALIDATION_ERROR", Message: message, HTTPStatus: http.StatusBadRequest}
}

func ErrNotFoundMsg(message string) *AppError {
	return &AppError{Code: "NOT_FOUND", Message: message, HTTPStatus: http.StatusNotFound}
}
