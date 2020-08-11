package postgres

import (
	"regexp"
)

var (
	ErrCodeIntegrity  = regexp.MustCompile("^ERROR #23000")
	ErrCodeRestrict   = regexp.MustCompile("^ERROR #23001")
	ErrCodeNotNull    = regexp.MustCompile("^ERROR #23502")
	ErrCodeForeignKey = regexp.MustCompile("^ERROR #23503")
	ErrCodeDuplicate  = regexp.MustCompile("^ERROR #23505")
)
