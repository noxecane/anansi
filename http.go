package anansi

import (
	"net/http"
	"strings"
)

// SimpleHeaders is a transformer for headers. It:
//
// - converts all header keys to lowercase
//
// - all single value headers to strings
func SimpleHeaders(headers http.Header) map[string]interface{} {
	lowerCaseHeaders := make(map[string]interface{})

	for k, v := range headers {
		lowerKey := strings.ToLower(k)
		if len(v) == 0 {
			lowerCaseHeaders[lowerKey] = ""
		} else if len(v) == 1 {
			lowerCaseHeaders[lowerKey] = v[0]
		} else {
			lowerCaseHeaders[lowerKey] = v
		}
	}

	return lowerCaseHeaders
}
