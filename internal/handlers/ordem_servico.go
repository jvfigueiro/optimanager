package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"optimanager/internal/database"
	"optimanager/internal/models"
	"optimanager/internal/services/pdf"
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
		respondError(w, http.StatusBadRequest, "Invalid service order ID")
		return
	}

	os, err := h.db.GetOrdemByID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if os == nil {
		respondError(w, http.StatusNotFound, "Service order not found")
		return
	}

	respondJSON(w, http.StatusOK, os)
}

func (h *OrdemHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateOrdemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON payload: "+err.Error())
		return
	}

	if strings.TrimSpace(req.Servico) == "" {
		respondError(w, http.StatusBadRequest, "Service description is required")
		return
	}
	if req.ClienteID <= 0 && strings.TrimSpace(req.ClienteNome) == "" {
		respondError(w, http.StatusBadRequest, "Client name or existing client ID is required")
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
		respondError(w, http.StatusBadRequest, "Invalid service order ID")
		return
	}

	var req models.UpdateOrdemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON payload: "+err.Error())
		return
	}

	if strings.TrimSpace(req.Servico) == "" {
		respondError(w, http.StatusBadRequest, "Service description is required")
		return
	}

	osAtualizada, err := h.db.UpdateOrdem(id, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, osAtualizada)
}

// BaixaPagamento settles payment status for an existing service order.
func (h *OrdemHandler) BaixaPagamento(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid service order ID")
		return
	}

	var req models.BaixaPagamentoRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	if req.Status == "" {
		req.Status = "Pago"
	}
	if req.Forma == "" {
		req.Forma = "Pix"
	}

	if err := h.db.UpdatePagamento(id, req.Forma, req.Status); err != nil {
		if strings.Contains(err.Error(), "not found") {
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

// DownloadPDF generates and streams native A4 PDF reports, quotes, or receipts via Maroto.
func (h *OrdemHandler) DownloadPDF(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid service order ID")
		return
	}

	tipo := r.URL.Query().Get("tipo")
	if tipo == "" {
		tipo = "relatorio"
	}

	os, err := h.db.GetOrdemByID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if os == nil {
		respondError(w, http.StatusNotFound, "Service order not found")
		return
	}

	cfg, err := h.db.GetConfig()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	pdfBytes, err := pdf.GenerateDocument(cfg, os, tipo)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "PDF generation failed: "+err.Error())
		return
	}

	filename := fmt.Sprintf("%s_OS_%d.pdf", strings.ToLower(tipo), id)
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filename))
	w.Header().Set("Content-Length", strconv.Itoa(len(pdfBytes)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdfBytes)
}
