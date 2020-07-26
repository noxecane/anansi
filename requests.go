package anansi

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

// ReadBody extracts the bytes in a request body without destroying the  contents of the body
func ReadBody(r *http.Request) []byte {
	var buffer bytes.Buffer

	// copy request body to in memory buffer while being read
	readSplit := io.TeeReader(r.Body, &buffer)
	body, err := ioutil.ReadAll(readSplit)
	if err != nil {
		panic(err)
	}

	if buffer.Len() == 0 {
		return make([]byte, 0)
	}

	// return what you collected
	r.Body = ioutil.NopCloser(&buffer)

	return body
}

// ReadBodyJSON is `ReadBody` except that it decodes
func ReadJSONBody(r *http.Request, schema ozzo.Validatable) {
	var buffer bytes.Buffer

	// copy request body to in memory buffer while being read
	readSplit := io.TeeReader(r.Body, &buffer)
	body, err := ioutil.ReadAll(readSplit)
	if err != nil {
		panic(err)
	}

	// return what you collected
	r.Body = ioutil.NopCloser(&buffer)

	content := r.Header.Get("Content-Type")
	if !strings.Contains(content, "application/json") || len(body) == 0 {
		// tell the user all the required attributes
		err = schema.Validate()
		if err != nil {
			panic(APIError{
				Code:    http.StatusBadRequest,
				Message: "We could not validate your request.",
				Meta:    err,
			})
		} else {
			return
		}
	}

	err = json.Unmarshal(body, &schema)
	if err != nil {
		panic(APIError{
			Code:    http.StatusBadRequest,
			Message: "We cannot understand your request.",
			Err:     err,
		})
	}

	// validate parsed JSON data
	err = schema.Validate()
	if err != nil {
		panic(APIError{
			Code:    http.StatusBadRequest,
			Message: "We could not validate your request.",
			Meta:    err,
		})
	}
}

// ReadJSONRaw is ReadJSON for arrays/maps. Was too lazy to generalise or name well.
func ReadJSONRaw(r *http.Request, schema interface{}) {
	var buffer bytes.Buffer

	// copy request body to in memory buffer while being read
	readSplit := io.TeeReader(r.Body, &buffer)
	body, err := ioutil.ReadAll(readSplit)
	if err != nil {
		panic(err)
	}

	// return what you collected
	r.Body = ioutil.NopCloser(&buffer)

	content := r.Header.Get("Content-Type")
	if !strings.Contains(content, "application/json") || len(body) == 0 {
		// tell the user all the required attributes
		err = ozzo.Validate(schema)
		if err != nil {
			panic(APIError{
				Code:    http.StatusBadRequest,
				Message: "We could not validate your request.",
				Meta:    err,
			})
		} else {
			return
		}
	}

	err = json.Unmarshal(body, &schema)
	if err != nil {
		panic(APIError{
			Code:    http.StatusBadRequest,
			Message: "We cannot understand your request.",
			Err:     err,
		})
	}

	// validate parsed JSON data
	err = ozzo.Validate(schema)
	if err != nil {
		panic(APIError{
			Code:    http.StatusBadRequest,
			Message: "We could not validate your request.",
			Meta:    err,
		})
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
