package postgres

import (
	"errors"
	"regexp"

	"github.com/go-pg/pg/v9"
)

// ErrDuplicate is the equivalent of 23505 integrity constraint violation
// for inserts and updates.
var ErrDuplicate = errors.New("row already exists")

var uniqueViolation = regexp.MustCompile("^ERROR #23505")

type Repo struct {
	*pg.DB
}

// DeduceError converts an error from pg methods(mainly insert and update)
// into an error type(like ErrDuplicate) so users don't have to depend on string
// interpretation to understand errors
func DeduceError(err error) error {
	// handle known errors
	switch {
	case uniqueViolation.MatchString(err.Error()):
		return ErrDuplicate
	default:
		return err
	}
}

// Create just calls Insert and handles known errors
func (r *Repo) Create(data interface{}) error {
	if err := r.DB.Insert(data); err != nil {
		return DeduceError(err)
	}

	return nil
}

// ByID returns the row with the primary key in the given data i.e ensure to set
// your primary key when passing your struct.
func (r *Repo) ByID(data interface{}) error {
	if err := r.DB.Model(data).WherePK().Select(); err != nil {
		return err
	}

	return nil
}

// UpdateOnly updates a row based on the primary key in the passed data, only
// changing the psql columns passed. Note that the ID of the struct must be set
// It returns the updated row when complete.
func (r *Repo) UpdateOnly(data interface{}, cols ...string) error {
	_, err := r.DB.
		Model(data).
		WherePK().
		Column(cols...).
		Returning("*").
		Update(data)

	if err != nil {
		return DeduceError(err)
	}

	return nil
}
