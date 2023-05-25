package cache

import (
	"bytes"
	"fmt"
)

type errorType int

const (
	ErrNotError errorType = iota
	ErrItemNotFound
	ErrCacheAlreadyRunning
)

var errMap = map[errorType]string{
	ErrNotError:            "not a valid error",
	ErrItemNotFound:        "item not found",
	ErrCacheAlreadyRunning: "cache already running",
}

func (e errorType) Error() string {
	return errMap[e]
}

func (e errorType) Is(target error) bool {
	t, ok := target.(errorType)
	if !ok {
		return false
	}
	return t == e
}

type integrityError struct {
	Errors []error
}

func NewIntegrityError(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	return &integrityError{
		Errors: errs,
	}
}

func (e *integrityError) Error() string {
	if len(e.Errors) == 0 {
		return ""
	}
	var errStrBuf bytes.Buffer
	errStrBuf.WriteString("Integrity errors have occurred:\n")
	for _, err := range e.Errors {
		errStrBuf.WriteString(fmt.Sprintf("\t%s\n", err))
	}
	return errStrBuf.String()
}

func IsIntegrityError(err error) bool {
	if err == nil {
		return false
	}
	var _, ok = err.(*integrityError)
	return ok
}
