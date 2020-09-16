package siber

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	ozzo "github.com/go-ozzo/ozzo-validation/v4"
)

// ReadBody extracts the bytes in a request body without destroying the contents of the body
func ReadBody(r *http.Request) []byte {
	var buffer bytes.Buffer

	// copy request body to in memory buffer while being read
	readSplit := io.TeeReader(r.Body, &buffer)
	body, err := ioutil.ReadAll(readSplit)
	if err != nil {
		panic(err)
	}

	// return what you collected
	r.Body = ioutil.NopCloser(&buffer)

	return body
}

// ReadJSON decodes the JSON body of the request and destroys to prevent possible issues with
// writing a response. If the content type is not JSON it fails with a 415. Otherwise it fails
// with a 400 on validation errors.
func ReadJSON(r *http.Request, v interface{}) {
	// make sure we are reading a JSON type
	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		panic(APIError{
			Code:    http.StatusUnsupportedMediaType,
			Message: http.StatusText(http.StatusUnsupportedMediaType),
		})
	}

	err := json.NewDecoder(r.Body).Decode(v)
	switch {
	case err == io.EOF:
		// tell the user all the required attributes
		err := ozzo.Validate(v)
		if err != nil {
			panic(APIError{
				Code:    http.StatusBadRequest,
				Message: "We could not validate your request.",
				Meta:    err,
			})
		}
		return
	case err != nil:
		panic(APIError{
			Code:    http.StatusBadRequest,
			Message: "We cannot parse your request body.",
			Err:     err,
		})
	default:
		// validate parsed JSON data
		err = ozzo.Validate(v)
		if err != nil {
			panic(APIError{
				Code:    http.StatusBadRequest,
				Message: "We could not validate your request.",
				Meta:    err,
			})
		}
	}
}

// IDParam extracts a uint URL parameter from the given request
func IDParam(r *http.Request, name string) uint {
	param := chi.URLParam(r, name)
	if param == "" {
		err := fmt.Sprintf("requested param %s is not part of route", name)
		panic(errors.New(err))
	}

	raw, err := strconv.ParseUint(param, 10, 32)
	if err != nil {
		panic(APIError{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("%s must be an ID", name),
		})
	}
	return uint(raw)
}

// StringParam basically just ensures the param name is correct. You might not
// need this method unless you're too lazy to do real tests.
func StringParam(r *http.Request, name string) string {
	param := chi.URLParam(r, name)
	if param == "" {
		err := fmt.Sprintf("requested param %s is not part of route", name)
		panic(errors.New(err))
	}

	return param
}
