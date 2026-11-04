package handlers

import (
	"encoding/json"
	"net/http"

	"optimanager/internal/database"
	"optimanager/internal/models"
)

type ConfigHandler struct {
	db *database.DB
}

func NewConfigHandler(db *database.DB) *ConfigHandler {
	return &ConfigHandler{db: db}
}

func (h *ConfigHandler) Get(w http.ResponseWriter, r *http.Request) {
	cfg, err := h.db.GetConfig()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, cfg)
}

func (h *ConfigHandler) Update(w http.ResponseWriter, r *http.Request) {
	var cfg models.Config
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON payload: "+err.Error())
		return
	}

	if err := h.db.UpdateConfig(&cfg); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	updated, err := h.db.GetConfig()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, updated)
}
