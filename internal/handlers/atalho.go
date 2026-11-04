package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"optimanager/internal/database"
	"optimanager/internal/models"
)

type AtalhoHandler struct {
	db *database.DB
}

func NewAtalhoHandler(db *database.DB) *AtalhoHandler {
	return &AtalhoHandler{db: db}
}

func (h *AtalhoHandler) List(w http.ResponseWriter, r *http.Request) {
	atalhos, err := h.db.ListAtalhos()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, atalhos)
}

func (h *AtalhoHandler) Create(w http.ResponseWriter, r *http.Request) {
	var a models.Atalho
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON payload: "+err.Error())
		return
	}

	if strings.TrimSpace(a.Titulo) == "" || strings.TrimSpace(a.Url) == "" {
		respondError(w, http.StatusBadRequest, "Shortcut title and URL are required")
		return
	}

	if err := h.db.CreateAtalho(&a); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, a)
}

func (h *AtalhoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid shortcut ID")
		return
	}

	if err := h.db.DeleteAtalho(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Shortcut deleted successfully"})
}
