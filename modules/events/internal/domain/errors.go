package domain

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
