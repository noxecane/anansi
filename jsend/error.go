package jsend

import "fmt"

// Err defines the structure of an error response that follows the JSend protocol
type Err struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Err     error       `json:"-"`
}

func (e Err) Error() string {
	if e.Err == nil {
		return e.Message
	}

	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

func (e Err) Unwrap() error { return e.Err }
