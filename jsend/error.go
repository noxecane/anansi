package jsend

import "fmt"

// Error defines the structure of an error response that follows the JSend protocol
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Err     error       `json:"-"`
}

func (e Error) Error() string {
	if e.Err == nil {
		return e.Message
	}

	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

func (e Error) Unwrap() error { return e.Err }
