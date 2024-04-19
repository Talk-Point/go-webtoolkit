package errors

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type ErrorDetail struct {
	Resource string `json:"resource"`
	Field    string `json:"field"`
	Value    string `json:"value"`
	Message  string `json:"message"`
}

type ErrorAlreadyExists struct {
	ErrorDetail
}

func (e *ErrorAlreadyExists) Error() string {
	return fmt.Sprintf("%s with %s %s already exists", e.Resource, e.Field, e.Value)
}

type ErrorNotFound struct {
	ErrorDetail
}

func (e *ErrorNotFound) Error() string {
	return fmt.Sprintf("%s with %s %s not found", e.Resource, e.Field, e.Value)
}

type ErrorBadRequest struct {
	ErrorDetail
}

func (e *ErrorBadRequest) Error() string {
	return "Bad request"
}

type ErrorSalechannelNotAllowed struct {
	ErrorDetail
}

func (e *ErrorSalechannelNotAllowed) Error() string {
	return fmt.Sprintf("Salechannel %s not allowed for you user", e.Value)
}

type ErrorUnauthorized struct {
	ErrorDetail
}

func (e *ErrorUnauthorized) Error() string {
	return "Unauthorized"
}

type ErrorUserNotActive struct {
	ErrorDetail
}

func (e *ErrorUserNotActive) Error() string {
	return "User not active"
}

type ErrorResponse struct {
	Message string        `json:"message"`
	Errors  []ErrorDetail `json:"errors"`
}

func validationErrorMessage(tag, param string) string {
	switch tag {
	case "required":
		return "is required."
	case "email":
		return "is not a valid email."
	case "eqfield":
		return "does not match the other field."
	case "min":
		return fmt.Sprintf("must be at least %s characters long.", param)
	case "max":
		return fmt.Sprintf("must be at most %s characters long.", param)
	case "uuid":
		return "is not a valid UUID."
	case "url":
		return "is not a valid URL."
	default:
		return "is not valid."
	}
}

func NewErrorResponse(err error) (*ErrorResponse, int) {
	switch e := err.(type) {
	case validator.ValidationErrors:
		var errs []ErrorDetail
		for _, err := range e {
			errs = append(errs, ErrorDetail{
				Resource: "Validation",
				Field:    err.Field(),
				Value:    "",
				Message:  fmt.Sprintf("%s %s", err.Field(), validationErrorMessage(err.Tag(), err.Param())),
			})
		}
		return &ErrorResponse{
			Message: "Validation failed",
			Errors:  errs,
		}, 400
	case *ErrorNotFound:
		return &ErrorResponse{
			Message: "Resource not found",
			Errors:  []ErrorDetail{e.ErrorDetail},
		}, 404
	case *ErrorBadRequest:
		return &ErrorResponse{
			Message: "Bad request",
			Errors:  []ErrorDetail{e.ErrorDetail},
		}, 400
	case *ErrorAlreadyExists:
		return &ErrorResponse{
			Message: "Resource already exists",
			Errors:  []ErrorDetail{e.ErrorDetail},
		}, 409
	case *ErrorSalechannelNotAllowed:
		return &ErrorResponse{
			Message: "Salechannel not allowed",
			Errors:  []ErrorDetail{e.ErrorDetail},
		}, 403
	case *ErrorUnauthorized:
		return &ErrorResponse{
			Message: "Unauthorized",
			Errors:  []ErrorDetail{e.ErrorDetail},
		}, 401
	case *ErrorUserNotActive:
		return &ErrorResponse{
			Message: "User not active",
			Errors:  []ErrorDetail{e.ErrorDetail},
		}, 401
	default:
		return &ErrorResponse{
			Message: "Internal server error",
			Errors: []ErrorDetail{
				{
					Resource: "Internal",
					Field:    "server",
					Value:    "error",
					Message:  err.Error(),
				},
			},
		}, 500
	}
}
