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

type OrdemHandler struct {
	db *database.DB
}

func NewOrdemHandler(db *database.DB) *OrdemHandler {
	return &OrdemHandler{db: db}
}

func (h *OrdemHandler) List(w http.ResponseWriter, r *http.Request) {
	filtro := r.URL.Query().Get("filtro")
	if filtro == "" {
		filtro = r.URL.Query().Get("status")
	}

	busca := r.URL.Query().Get("busca")
	if busca == "" {
		busca = r.URL.Query().Get("search")
	}

	ordens, err := h.db.ListOrdens(filtro, busca)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, ordens)
}

func (h *OrdemHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "ID da ordem de serviço inválido")
		return
	}

	os, err := h.db.GetOrdemByID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if os == nil {
		respondError(w, http.StatusNotFound, "Ordem de serviço não encontrada")
		return
	}

	respondJSON(w, http.StatusOK, os)
}

func (h *OrdemHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateOrdemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Payload JSON inválido: "+err.Error())
		return
	}

	if strings.TrimSpace(req.Servico) == "" {
		respondError(w, http.StatusBadRequest, "A descrição do serviço é obrigatória")
		return
	}
	if req.ClienteID <= 0 && strings.TrimSpace(req.ClienteNome) == "" {
		respondError(w, http.StatusBadRequest, "Informe o nome do cliente ou o ID de um cliente existente")
		return
	}

	novaOS, err := h.db.CreateOrdem(&req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, novaOS)
}

func (h *OrdemHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "ID da ordem de serviço inválido")
		return
	}

	var req models.UpdateOrdemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Payload JSON inválido: "+err.Error())
		return
	}

	if strings.TrimSpace(req.Servico) == "" {
		respondError(w, http.StatusBadRequest, "A descrição do serviço é obrigatória")
		return
	}

	osAtualizada, err := h.db.UpdateOrdem(id, &req)
	if err != nil {
		if strings.Contains(err.Error(), "não encontrada") {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, osAtualizada)
}

// BaixaPagamento liquida rapidamente o pagamento de uma OS pendente.
func (h *OrdemHandler) BaixaPagamento(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "ID da ordem de serviço inválido")
		return
	}

	var req models.BaixaPagamentoRequest
	// Corpo opcional (se não vier, default é Pix / Pago)
	_ = json.NewDecoder(r.Body).Decode(&req)

	if req.Status == "" {
		req.Status = "Pago"
	}
	if req.Forma == "" {
		req.Forma = "Pix"
	}

	if err := h.db.UpdatePagamento(id, req.Forma, req.Status); err != nil {
		if strings.Contains(err.Error(), "não encontrada") {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	os, err := h.db.GetOrdemByID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, os)
}

