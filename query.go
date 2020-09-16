package siber

import (
	"net/http"
	"reflect"
	"strconv"

	"github.com/pkg/errors"
)

// ParseQuery converts a map of strings to strings to a struct. It uses
// the key and default struct tag to help it determine how to get keys and set defaults.
// Mind that it only supports the types
// - int(all bit sizes)
// - uint(all bit sizes)
// - float(all bit sizes)
// - string
// - ISO8601 time
func ParseQuery(query map[string]string, v interface{}) error {
	result := reflect.ValueOf(v).Elem()
	resultType := result.Type()

	for i := 0; i < result.NumField(); i++ {
		field := resultType.Field(i)
		fieldVal := result.Field(i)
		fieldType := field.Type
		fieldKind := fieldVal.Kind()

		// skip hidden fields
		if field.PkgPath != "" {
			continue
		}

		// get query parameter name
		key := field.Tag.Get("key")
		if key == "" {
			key = field.Name
		}

		// for fields with default values
		def := field.Tag.Get("default")

		rawValue, ok := query[key]
		if !ok {
			if def == "" {
				fieldVal.Set(reflect.Zero(fieldType))
				continue
			} else {
				rawValue = def
			}
		}

		// make sure we're not using a pointer
		var ptr reflect.Value
		if fieldKind == reflect.Ptr {
			// get the underlying type of the pointer
			fieldType = fieldType.Elem()
			fieldKind = fieldType.Kind()

			// create new pointer to hold the value
			ptr = reflect.New(fieldType)
		}

		if !fieldVal.CanSet() {
			return errors.Errorf("cannot set field %s", field.Name)
		}

		var out interface{}
		var err error

		switch fieldKind {
		case reflect.Bool:
			out, err = strconv.ParseBool(rawValue)
			if err != nil {
				return errors.Wrapf(err, "failed to parse bool %s", field.Name)
			}
		case reflect.Int:
			i, err := strconv.ParseInt(rawValue, 10, fieldType.Bits())
			if err != nil {
				return errors.Wrapf(err, "failed to parse int %s", field.Name)
			}
			out = int(i)
		case reflect.Int8:
			i, err := strconv.ParseInt(rawValue, 10, 8)
			if err != nil {
				return errors.Wrapf(err, "failed to parse int8 %s", field.Name)
			}
			out = int8(i)
		case reflect.Int16:
			i, err := strconv.ParseInt(rawValue, 10, 16)
			if err != nil {
				return errors.Wrapf(err, "failed to parse int16 %s", field.Name)
			}
			out = int16(i)
		case reflect.Int32:
			i, err := strconv.ParseInt(rawValue, 10, 32)
			if err != nil {
				return errors.Wrapf(err, "failed to parse int32 %s", field.Name)
			}
			out = int32(i)
		case reflect.Int64:
			if out, err = strconv.ParseInt(rawValue, 10, 64); err != nil {
				return errors.Wrapf(err, "failed to parse int64 %s", field.Name)
			}
		case reflect.Uint:
			u, err := strconv.ParseUint(rawValue, 10, fieldType.Bits())
			if err != nil {
				return errors.Wrapf(err, "failed to parse int %s", field.Name)
			}
			out = uint(u)
		case reflect.Uint8:
			u, err := strconv.ParseUint(rawValue, 10, 8)
			if err != nil {
				return errors.Wrapf(err, "failed to parse int8 %s", field.Name)
			}
			out = uint8(u)
		case reflect.Uint16:
			u, err := strconv.ParseUint(rawValue, 10, 16)
			if err != nil {
				return errors.Wrapf(err, "failed to parse int16 %s", field.Name)
			}
			out = uint16(u)
		case reflect.Uint32:
			u, err := strconv.ParseUint(rawValue, 10, 32)
			if err != nil {
				return errors.Wrapf(err, "failed to parse int32 %s", field.Name)
			}
			out = uint32(u)
		case reflect.Uint64:
			if out, err = strconv.ParseUint(rawValue, 10, 64); err != nil {
				return errors.Wrapf(err, "failed to parse int64 %s", field.Name)
			}
		case reflect.Float32:
			f, err := strconv.ParseFloat(rawValue, fieldType.Bits())
			if err != nil {
				return errors.Wrapf(err, "failed to parse float %s", field.Name)
			}
			out = float32(f)
		case reflect.Float64:
			if out, err = strconv.ParseFloat(rawValue, fieldType.Bits()); err != nil {
				return errors.Wrapf(err, "failed to parse float %s", field.Name)
			}
		case reflect.String:
			out = rawValue
		case reflect.Struct:
			// attempt to parse a date value
			if out, err = FromISO(rawValue); err != nil {
				return errors.Errorf("this function doesn't support %s for the field '%s'", fieldType, field.Name)
			}
		default:
			return errors.Errorf("this function doesn't support %s for the field '%s'", fieldType, field.Name)
		}

		// if original kind is pointer, save as pointer value
		if fieldVal.Kind() == reflect.Ptr {
			// set value pointer is pointing to
			reflect.Indirect(ptr).Set(reflect.ValueOf(out))
			fieldVal.Set(ptr)
		} else {
			fieldVal.Set(reflect.ValueOf(out))
		}
	}

	return nil
}

// ReadQuery reads the query parameters of a request into a struct
func ReadQuery(r *http.Request, v interface{}) {
	raw := r.URL.Query()
	qMap := make(map[string]string)

	for k := range raw {
		qMap[k] = raw.Get(k)
	}

	if err := ParseQuery(qMap, v); err != nil {
		panic(JSendError{
			Code:    http.StatusBadRequest,
			Message: "We could not parse your request query.",
			Err:     err,
		})
	}
}
