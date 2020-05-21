package anansi

import "fmt"

// APIError is a struct describing an error
type APIError struct {
	Code    int         `json:"-"`
	Message string      `json:"message"`
	Meta    interface{} `json:"meta"`
}

type ErrorInterpreter func(err error) APIError

// implements the error interface
func (e APIError) Error() string { return e.Message }

func ConstraintError(message string) APIError {
	return APIError{
		Code:    422,
		Message: message,
	}
}

func ForbiddenError(message string) APIError {
	return APIError{
		Code:    403,
		Message: message,
	}
}

func UnauthorisedError(message string) APIError {
	return APIError{
		Code:    401,
		Message: message,
	}
}

func BadRequestError(message string) APIError {
	return APIError{
		Code:    400,
		Message: message,
	}
}

func ConflictError(entity string) APIError {
	return APIError{
		Code:    409,
		Message: fmt.Sprintf("There's already an existing %s", entity),
	}
}

func BadRequestDataError(message string, data interface{}) APIError {
	return APIError{
		Code:    400,
		Message: message,
		Meta:    data,
	}
}
