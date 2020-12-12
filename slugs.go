package siber

// Again stealing from those that have sense
// https://github.com/kelseyhightower/envconfig/blob/master/envconfig.go

import (
	"regexp"
	"strings"
)

var byCaseRegexp = regexp.MustCompile("([^A-Z]+|[A-Z]+[^A-Z]+|[A-Z]+)")
var acronymRegexp = regexp.MustCompile("([A-Z]+)([A-Z][^A-Z]+)")

func Slugify(s string) string {
	// split by captialised words
	words := byCaseRegexp.FindAllStringSubmatch(s, -1)

	// not really our business
	if len(words) == 0 {
		return s
	}

	var slug []string
	for _, match := range words {
		// extract acronym and capitalised word
		acrs := acronymRegexp.FindStringSubmatch(match[1])
		if len(acrs) == 3 {
			slug = append(slug, strings.ToLower(acrs[1]), strings.ToLower(acrs[2]))
		} else {
			slug = append(slug, strings.ToLower(match[1]))
		}
	}

	return strings.Join(slug, "_")
}
