package siber

import "fmt"

// APIError is a struct describing an error
type APIError struct {
	Code    int         `json:"-"`
	Message string      `json:"message"`
	Meta    interface{} `json:"meta"`
	Err     error       `json:"-"`
}

// implements the error interface
func (e APIError) Error() string { return fmt.Sprintf("%s: %v", e.Message, e.Err) }

func (e APIError) Unwrap() error { return e.Err }
