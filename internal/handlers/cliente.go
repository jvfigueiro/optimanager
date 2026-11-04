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

type ClienteHandler struct {
	db *database.DB
}

func NewClienteHandler(db *database.DB) *ClienteHandler {
	return &ClienteHandler{db: db}
}

func (h *ClienteHandler) List(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	if search == "" {
		search = r.URL.Query().Get("q")
	}

	clientes, err := h.db.ListClientes(search)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, clientes)
}

func (h *ClienteHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid client ID")
		return
	}

	cliente, err := h.db.GetClienteByID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if cliente == nil {
		respondError(w, http.StatusNotFound, "Client not found")
		return
	}

	respondJSON(w, http.StatusOK, cliente)
}

func (h *ClienteHandler) Create(w http.ResponseWriter, r *http.Request) {
	var c models.Cliente
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON payload: "+err.Error())
		return
	}

	if strings.TrimSpace(c.Nome) == "" {
		respondError(w, http.StatusBadRequest, "Client name is required")
		return
	}

	if err := h.db.CreateCliente(&c); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, c)
}

func (h *ClienteHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid client ID")
		return
	}

	var c models.Cliente
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON payload: "+err.Error())
		return
	}
	c.ID = id

	if strings.TrimSpace(c.Nome) == "" {
		respondError(w, http.StatusBadRequest, "Client name is required")
		return
	}

	if err := h.db.UpdateCliente(&c); err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		if strings.Contains(err.Error(), "already exists") {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, c)
}

func (h *ClienteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid client ID")
		return
	}

	if err := h.db.DeleteCliente(id); err != nil {
		if strings.Contains(err.Error(), "cannot delete client") {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Client deleted successfully"})
}
