package models

import "time"

type AuditLog struct {
	ID          int64     `json:"id"`
	UsuarioID   int64     `json:"usuario_id"`
	UsuarioNome string    `json:"usuario_nome"`
	IP          string    `json:"ip"`
	Acao        string    `json:"acao"`
	Detalhe     string    `json:"detalhe"`
	CreatedAt   time.Time `json:"created_at"`
}

