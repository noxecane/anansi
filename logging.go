package siber

import (
	"os"
	"regexp"

	"github.com/rs/zerolog"
)

var DumpLog = regexp.MustCompile("(?i)kube-probe|prometheus")

func NewLogger(service string) zerolog.Logger {
	host, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	return zerolog.
		New(os.Stdout).
		With().
		Timestamp().
		Str("service", service).
		Str("host", host).
		Logger()
}
