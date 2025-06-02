package liberrors

import "github.com/go-errors/errors"

func NewErrorf(msg string, a ...interface{}) error {
	err := errors.Errorf(msg, a...)
	return errors.New(err)
}