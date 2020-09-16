package siber

import "fmt"

// JSendError is a struct describing an error
type JSendError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Err     error       `json:"-"`
}

// implements the error interface
func (e JSendError) Error() string { return fmt.Sprintf("%s: %v", e.Message, e.Err) }

func (e JSendError) Unwrap() error { return e.Err }
