package api

import (
	"net/http"

	"github.com/noxecane/anansi/sessions"
)

func LoadBearer(m *sessions.Manager, r *http.Request, v interface{}) {
	err := m.LoadBearer(r, v)
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
			Err:     err,
		})
	}
}

func LoadCookie(m *sessions.Manager, r *http.Request, v interface{}) {
	err := m.LoadCookie(r, v)
	if err == nil {
		return
	}

	switch err {
	case sessions.ErrEmptyAuthCookie:
		panic(Err{
			Code:    http.StatusUnauthorized,
			Message: "Your request is not authenticated",
		})
	default:
		panic(Err{
			Code:    http.StatusUnauthorized,
			Message: "Your token is either invalid or has expired",
			Err:     err,
		})
	}
}

func LoadHeadless(m *sessions.Manager, r *http.Request, v interface{}) {
	err := m.LoadHeadless(r, v)
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
			Err:     err,
		})
	}
}

func Load(m *sessions.Manager, r *http.Request, v interface{}) {
	err := m.Load(r, v)
	if err == nil {
		return
	}

	switch err {
	case sessions.ErrEmptyAuthCookie:
		panic(Err{
			Code:    http.StatusUnauthorized,
			Message: "Your request is not authenticated",
		})
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
			Err:     err,
		})
	}
}
