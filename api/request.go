package api

import (
	"errors"
	"fmt"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/noxecane/anansi/requests"
)

// ReadJSON parses the body of an http request into the value pointed by v.
// It panics with a 415 error if the content type is not JSON, a 400 if the value fails
// ozzo validation or any other error(JSON decode error for instance)
func ReadJSON(r *http.Request, v interface{}) {
	err := requests.ReadJSON(r, v)
	if err == nil {
		return
	}

	var e validation.Errors
	switch {
	case err == requests.ErrNotJSON:
		panic(Err{
			Code:    http.StatusUnsupportedMediaType,
			Message: http.StatusText(http.StatusUnsupportedMediaType),
			Err:     err,
		})
	case errors.As(err, &e):
		panic(Err{
			Code:    http.StatusBadRequest,
			Message: "We could not validate your request.",
			Data:    e,
		})
	default:
		panic(Err{
			Code:    http.StatusBadRequest,
			Message: "We cannot parse your request body.",
			Err:     err,
		})
	}
}

// QueryParams converts the query values of the request into a struct using
// the "json" tag to map the keys. It supports transformations using modl
// and validation provided by ozzo. It panics with a 400 if the value fails
// ozzo validation or any other error
func QueryParam(r *http.Request, v interface{}) {
	err := requests.QueryParams(r, v)
	if err == nil {
		return
	}

	var e validation.Errors
	switch {
	case errors.As(err, &e):
		panic(Err{
			Code:    http.StatusBadRequest,
			Message: "We could not validate your request.",
			Data:    e,
		})
	default:
		panic(Err{
			Code:    http.StatusBadRequest,
			Message: "We cannot parse your request body.",
			Err:     err,
		})
	}
}

// IDParam extracts a uint URL parameter from the given request. panics with a 400 if
// the param is not a strin, otherwise it panics with a basic error.
func IDParam(r *http.Request, name string) uint {
	id, err := requests.IDParam(r, name)
	if err != nil {
		panic(Err{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("%s must be an integer ID", name),
		})
	}

	return id
}

// StringParam basically just ensures the param name is correct. You might not
// need this method unless you're too lazy to do real tests. Panics if there's no param
// with the name.
var StringParam = requests.StringParam
