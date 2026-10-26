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
		respondError(w, http.StatusBadRequest, "ID de cliente inválido")
		return
	}

	cliente, err := h.db.GetClienteByID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if cliente == nil {
		respondError(w, http.StatusNotFound, "Cliente não encontrado")
		return
	}

	respondJSON(w, http.StatusOK, cliente)
}

func (h *ClienteHandler) Create(w http.ResponseWriter, r *http.Request) {
	var c models.Cliente
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		respondError(w, http.StatusBadRequest, "Payload JSON inválido: "+err.Error())
		return
	}

	if strings.TrimSpace(c.Nome) == "" {
		respondError(w, http.StatusBadRequest, "O nome do cliente é obrigatório")
		return
	}

	if err := h.db.CreateCliente(&c); err != nil {
		if strings.Contains(err.Error(), "já existe") {
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
		respondError(w, http.StatusBadRequest, "ID de cliente inválido")
		return
	}

	var c models.Cliente
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		respondError(w, http.StatusBadRequest, "Payload JSON inválido: "+err.Error())
		return
	}
	c.ID = id

	if strings.TrimSpace(c.Nome) == "" {
		respondError(w, http.StatusBadRequest, "O nome do cliente é obrigatório")
		return
	}

	if err := h.db.UpdateCliente(&c); err != nil {
		if strings.Contains(err.Error(), "não encontrado") {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		if strings.Contains(err.Error(), "já existe") {
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
		respondError(w, http.StatusBadRequest, "ID de cliente inválido")
		return
	}

	if err := h.db.DeleteCliente(id); err != nil {
		if strings.Contains(err.Error(), "não é possível excluir") {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		if strings.Contains(err.Error(), "não encontrado") {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Cliente excluído com sucesso"})
}

