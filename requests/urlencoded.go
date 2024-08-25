package requests

import (
	"context"
	"errors"
	"net/http"

	ozzo "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-viper/mapstructure/v2"
	"github.com/noxecane/anansi"
)

// QueryParams converts the query values of the request into a struct using
// the "json" tag to map the keys. It supports transformations using modl
// and validation provided by ozzo.
func QueryParams(r *http.Request, v interface{}) error {
	return parseParams(r.Context(), r.URL.Query(), v)
}

// FormData is QueryParams for x-www-form-urlencoded
func FormData(r *http.Request, v interface{}) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	return parseParams(r.Context(), r.Form, v)
}

// MultipartFormData converts the non-files in the form data of the request
// into a struct using the "json" tag to map the keys. It supports
// transformations using modl and validation provided by ozzo.
func MultipartFormData(r *http.Request, size int64, v interface{}) error {
	err := r.ParseMultipartForm(size)
	if err != nil {
		return err
	}

	return parseParams(r.Context(), r.MultipartForm.Value, v)
}

func parseParams(ctx context.Context, values map[string][]string, v interface{}) error {
	params := anansi.SimpleMap(values)

	config := &mapstructure.DecoderConfig{Result: v, TagName: `json`, WeaklyTypedInput: true}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return errors.Join(err, errors.New("could not convert query parameters to struct"))
	}

	err = decoder.Decode(params)
	if err != nil {
		return err
	}

	// validate parsed JSON data
	if err := generalMold.Struct(ctx, v); err != nil {
		return err
	}

	return ozzo.Validate(v)
}
