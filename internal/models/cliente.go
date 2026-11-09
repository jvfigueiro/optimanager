package models

import "time"

type Cliente struct {
	ID        int64     `json:"id"`
	Nome      string    `json:"nome"`
	Doc       string    `json:"doc"`
	Contato   string    `json:"contato"`
	DesdeAno  int       `json:"desde_ano"`
	CreatedAt time.Time `json:"created_at"`
}

// BatchClientesRequest defines payload for bulk operations on clients.
type BatchClientesRequest struct {
	Action string  `json:"action"` // "delete"
	IDs    []int64 `json:"ids"`
}

// BatchClientesResponse returns details of bulk client deletion.
type BatchClientesResponse struct {
	Deleted []int64 `json:"deleted"`
	Failed  []int64 `json:"failed"`
	Message string  `json:"message"`
}

