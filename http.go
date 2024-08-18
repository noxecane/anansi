package anansi

import (
	"strings"
)

// SimpleMap flattens a map of string arrays. Useful for headers and query
// values. Note that it converts all keys to lowercase
func SimpleMap(m map[string][]string) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range m {
		lowerKey := strings.ToLower(k)
		if len(v) == 0 {
			result[lowerKey] = ""
		} else if len(v) == 1 {
			result[lowerKey] = v[0]
		} else {
			result[lowerKey] = v
		}
	}

	return result
}
