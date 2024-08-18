package requests

import (
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
	params := anansi.SimpleMap(r.URL.Query())

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
	if err := generalMold.Struct(r.Context(), v); err != nil {
		return err
	}

	return ozzo.Validate(v)
}
