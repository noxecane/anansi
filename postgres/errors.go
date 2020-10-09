package postgres

import (
	"regexp"
)

var (
	ErrIntegrity  = regexp.MustCompile("^ERROR #23000")
	ErrRestrict   = regexp.MustCompile("^ERROR #23001")
	ErrNotNull    = regexp.MustCompile("^ERROR #23502")
	ErrForeignKey = regexp.MustCompile("^ERROR #23503")
	ErrDuplicate  = regexp.MustCompile("^ERROR #23505")
)
