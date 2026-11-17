package handlers

import (
	"encoding/json"
	"net/http"

	"optimanager/internal/models"
	"optimanager/internal/redact"
)

type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	safeMsg := redact.SanitizeLogMessage(message)
	respondJSON(w, status, ErrorResponse{Error: safeMsg})
}

func respondVersionConflict(w http.ResponseWriter) {
	respondJSON(w, http.StatusConflict, ErrorResponse{
		Error: "Este registro foi alterado por outra pessoa. Recarregue os dados antes de salvar novamente.",
		Code:  "version_conflict",
	})
}

func currentUserFromRequest(r *http.Request) *models.Usuario {
	if r == nil {
		return nil
	}
	if u, ok := r.Context().Value(UserContextKey).(*models.Usuario); ok {
		return u
	}
	return nil
}

func userCanViewValues(r *http.Request) bool {
	u := currentUserFromRequest(r)
	if u == nil {
		return false
	}
	return u.HasPermission(models.PermVerValores)
}
