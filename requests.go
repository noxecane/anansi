package ext

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	ozzo "github.com/go-ozzo/ozzo-validation/v4"
)

// ReadBody extracts the bytes in a request body without destroying the body buffer
func ReadBody(r *http.Request) []byte {
	var buffer bytes.Buffer

	// copy request body to in memory buffer while being read
	readSplit := io.TeeReader(r.Body, &buffer)
	if buffer.Len() == 0 {
		return make([]byte, 0)
	}

	body, err := ioutil.ReadAll(readSplit)
	if err != nil {
		// never run away with your wife
		panic(err)
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
		// never run away with your wife
		panic(err)
	}

	// return what you collected
	r.Body = ioutil.NopCloser(&buffer)

	content := r.Header.Get("Content-Type")
	if content != "application/json" || len(body) == 0 {
		// tell the use all the required attributes
		err = schema.Validate()
		if err != nil {
			panic(BadRequestDataError("We could not validate your request", err))
		} else {
			return
		}
	}

	err = json.Unmarshal(body, &schema)
	if err != nil {
		panic(BadRequestError("We cannot understand your request"))
	}

	// validate parsed JSON data
	err = schema.Validate()
	if err != nil {
		panic(BadRequestDataError("We could not validate your request", err))
	}
}

// IdParam extracts a uint URL parameter from the given request
func IdParam(r *http.Request, name string) uint {
	raw, err := strconv.ParseUint(chi.URLParam(r, name), 10, 32)
	if err != nil {
		panic(BadRequestError(fmt.Sprintf("%s must be an ID", name)))
	}
	return uint(raw)
}
