package api

import "fmt"

// Err defines the structure of an HTTP error response
type Err struct {
	Code    int         `json:"-"`
	Message string      `json:"message"`
	Data    interface{} `json:"meta"`
	Err     error       `json:"-"`
}

func (e Err) Error() string {
	if e.Err == nil {
		return e.Message
	}

	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

func (e Err) Unwrap() error { return e.Err }
