package domain

import "net/http"

type ErrorKind int

const (
	KindValidation ErrorKind = iota
	KindNotFound
	KindConflict
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
	default:
		return http.StatusBadRequest
	}
}

var (
	ErrUserNotFound      = &DomainError{Code: "user.not_found", Message: "user not found", Kind: KindNotFound}
	ErrEmailAlreadyTaken = &DomainError{Code: "user.email_taken", Message: "email is already in use", Kind: KindConflict}
)
