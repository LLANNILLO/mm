package presentation

import (
	"errors"
	"net/http"

	"github.com/llannillo/mm/modules/events/internal/domain"
)

func handleDomainError(w http.ResponseWriter, err error) {
	var de *domain.DomainError
	if errors.As(err, &de) {
		switch de.Kind {
		case domain.KindValidation:
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": de.Message, "code": de.Code})
			return
		case domain.KindNotFound:
			writeJSON(w, http.StatusNotFound, map[string]string{"error": de.Message, "code": de.Code})
			return
		case domain.KindConflict:
			writeJSON(w, http.StatusConflict, map[string]string{"error": de.Message, "code": de.Code})
			return
		}
	}
	writeError(w, http.StatusInternalServerError, "internal server error")
}
