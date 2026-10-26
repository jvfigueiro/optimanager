package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"optimanager/internal/database"
)

type HealthResponse struct {
	Status      string    `json:"status"`
	App         string    `json:"app"`
	Version     string    `json:"version"`
	Database    string    `json:"database"`
	JournalMode string    `json:"journal_mode"`
	ForeignKeys int       `json:"foreign_keys"`
	Timestamp   time.Time `json:"timestamp"`
}

type HealthHandler struct {
	db *database.DB
}

func NewHealthHandler(db *database.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	var journalMode string
	var foreignKeys int

	// Testa a conexão do pool de leitura
	err1 := h.db.Reader.QueryRow("PRAGMA journal_mode;").Scan(&journalMode)
	err2 := h.db.Reader.QueryRow("PRAGMA foreign_keys;").Scan(&foreignKeys)

	dbStatus := "connected"
	if err1 != nil || err2 != nil {
		dbStatus = "error"
	}

	resp := HealthResponse{
		Status:      "ok",
		App:         "OptiManager",
		Version:     "3.0.0",
		Database:    dbStatus,
		JournalMode: journalMode,
		ForeignKeys: foreignKeys,
		Timestamp:   time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	if dbStatus != "connected" {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	_ = json.NewEncoder(w).Encode(resp)
}

