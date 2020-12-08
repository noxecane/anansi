package jsend

import (
	"net/http"

	"github.com/random-guys/go-siber/sessions"
)

func LoadBearer(s *sessions.Store, r *http.Request, v interface{}) {
	err := s.LoadBearer(r, v)
	if err == nil {
		return
	}

	switch err {
	case sessions.ErrEmptyHeader:
		panic(Err{
			Code:    http.StatusUnauthorized,
			Message: "Your request is not authenticated",
		})
	case sessions.ErrHeaderFormat:
		panic(Err{
			Code:    http.StatusUnauthorized,
			Message: "Your authorization header is incorrect",
		})
	case sessions.ErrUnsupportedScheme:
		panic(Err{
			Code:    http.StatusUnauthorized,
			Message: "We don't support your authorization scheme",
		})
	default:
		panic(Err{
			Code:    http.StatusUnauthorized,
			Message: "Your token is either invalid or has expired",
		})
	}
}

func LoadHeadless(s *sessions.Store, r *http.Request, v interface{}) {
	err := s.LoadHeadless(r, v)
	if err == nil {
		return
	}

	switch err {
	case sessions.ErrEmptyHeader:
		panic(Err{
			Code:    http.StatusUnauthorized,
			Message: "Your request is not authenticated",
		})
	case sessions.ErrHeaderFormat:
		panic(Err{
			Code:    http.StatusUnauthorized,
			Message: "Your authorization header is incorrect",
		})
	case sessions.ErrUnsupportedScheme:
		panic(Err{
			Code:    http.StatusUnauthorized,
			Message: "We don't support your authorization scheme",
		})
	default:
		panic(Err{
			Code:    http.StatusUnauthorized,
			Message: "Your token is either invalid or has expired",
		})
	}
}

func Load(s *sessions.Store, r *http.Request, v interface{}) {
	err := s.Load(r, v)
	if err == nil {
		return
	}

	switch err {
	case sessions.ErrEmptyHeader:
		panic(Err{
			Code:    http.StatusUnauthorized,
			Message: "Your request is not authenticated",
		})
	case sessions.ErrHeaderFormat:
		panic(Err{
			Code:    http.StatusUnauthorized,
			Message: "Your authorization header is incorrect",
		})
	case sessions.ErrUnsupportedScheme:
		panic(Err{
			Code:    http.StatusUnauthorized,
			Message: "We don't support your authorization scheme",
		})
	default:
		panic(Err{
			Code:    http.StatusUnauthorized,
			Message: "Your token is either invalid or has expired",
		})
	}
}
