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

