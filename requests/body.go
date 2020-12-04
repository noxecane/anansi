package requests

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	ozzo "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/random-guys/go-siber/json"
)

var ErrNotJSON = errors.New("body is not JSON")

// ReadBody extracts the bytes in a request body without destroying the contents of the body.
// Returns an error if reading body fails.
func ReadBody(r *http.Request) ([]byte, error) {
	var buffer bytes.Buffer

	// copy request body to in memory buffer while being read
	readSplit := io.TeeReader(r.Body, &buffer)
	body, err := ioutil.ReadAll(readSplit)
	if err != nil {
		return nil, err
	}

	// return what you collected
	r.Body = ioutil.NopCloser(&buffer)

	return body, nil
}

// ReadJSON decodes the JSON body of the request and destroys it to prevent possible issues with
// writing a response. Returns ErrNotJSON if the content-type of the request is not JSON, else
// it returns validation.Errors if the resultant value fails validation defined using ozzo.
// Otherwise the it returns an error when json decoding fails
func ReadJSON(r *http.Request, v interface{}) error {
	// make sure we are reading a JSON type
	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		return ErrNotJSON
	}

	err := json.NewDecoder(r.Body).Decode(v)
	switch {
	case err == io.EOF:
		// tell the user all the required attributes
		err := ozzo.Validate(v)
		if err != nil {
			return err
		}
		return err
	case err != nil:
		return err
	default:
		// validate parsed JSON data
		err = ozzo.Validate(v)
		if err != nil {
			return err
		}
	}

	return nil
}

// IDParam extracts a uint URL parameter from the given request. panics if there's no
// such path on the route, otherwise it returns an error if the param is not an int.
func IDParam(r *http.Request, name string) (uint, error) {
	param := chi.URLParam(r, name)
	if param == "" {
		err := fmt.Sprintf("requested param %s is not part of route", name)
		panic(errors.New(err))
	}

	raw, err := strconv.ParseUint(param, 10, 32)
	if err != nil {
		err := fmt.Sprintf("%s must be an ID", name)
		return 0, errors.New(err)
	}

	return uint(raw), nil
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
