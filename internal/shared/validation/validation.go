package validation

import "strings"

type Failure struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrors struct {
	Failures []Failure
}

func (e *ValidationErrors) Error() string { return "one or more validation errors occurred" }

type Builder struct {
	failures []Failure
}

func New() *Builder { return &Builder{} }

// Required adds a failure if value is empty or whitespace-only.
func (b *Builder) Required(field, value string) *Builder {
	if strings.TrimSpace(value) == "" {
		b.failures = append(b.failures, Failure{Field: field, Message: field + " is required"})
	}
	return b
}

// Custom adds a failure when failed is true.
func (b *Builder) Custom(field string, failed bool, message string) *Builder {
	if failed {
		b.failures = append(b.failures, Failure{Field: field, Message: message})
	}
	return b
}

// Err returns nil when there are no failures, or *ValidationErrors.
func (b *Builder) Err() error {
	if len(b.failures) == 0 {
		return nil
	}
	return &ValidationErrors{Failures: b.failures}
}
