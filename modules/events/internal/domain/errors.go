package domain

import "net/http"

type ErrorKind int

const (
	KindValidation ErrorKind = iota
	KindNotFound
	KindConflict
	KindProblem
)

type DomainError struct {
	Code    string
	Message string
	Kind    ErrorKind
}

func (e *DomainError) Error() string { return e.Message }

func (e *DomainError) StatusCode() int {
	switch e.Kind {
	case KindNotFound:
		return http.StatusNotFound
	case KindConflict:
		return http.StatusConflict
	case KindValidation, KindProblem:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
