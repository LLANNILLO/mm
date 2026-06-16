package handler

import (
	"errors"
	"net/http"

	"github.com/llannillo/mm/internal/shared/problem"
	"github.com/llannillo/mm/internal/shared/validation"
	"github.com/llannillo/mm/modules/events/internal/domain"
)

func handleDomainError(w http.ResponseWriter, err error) {
	var ve *validation.ValidationErrors
	if errors.As(err, &ve) {
		problem.Write(w, problem.Detail{
			Title:  "Validation failure",
			Status: http.StatusBadRequest,
			Detail: "One or more validation errors occurred.",
			Errors: ve.Failures,
		})
		return
	}

	var de *domain.DomainError
	if !errors.As(err, &de) {
		problem.WriteInternal(w)
		return
	}
	problem.Write(w, problem.Detail{
		Title:  de.Code,
		Status: de.StatusCode(),
		Detail: de.Message,
	})
}
