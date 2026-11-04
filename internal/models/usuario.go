package models

import "time"

type Usuario struct {
	ID        int64     `json:"id"`
	Nome      string    `json:"nome"`
	Email     string    `json:"email"`
	SenhaHash string    `json:"-"`
	Role      string    `json:"role"` // 'admin' | 'operador'
	CreatedAt time.Time `json:"created_at"`
}

type SetupAdminRequest struct {
	// Dados do Administrador
	Nome  string `json:"nome"`
	Email string `json:"email"`
	Senha string `json:"senha"`

	// Dados da Empresa (Configurados no Assistente de Primeiro Acesso)
	Fantasia     string `json:"fantasia"`
	RazaoSocial  string `json:"razao_social"`
	CNPJ         string `json:"cnpj"`
	Fones        string `json:"fones"`
	EmailContato string `json:"email_contato"`
	Endereco     string `json:"endereco"`
	Cidade       string `json:"cidade"`
}

type LoginRequest struct {
	Email string `json:"email"`
	Senha string `json:"senha"`
}

type CreateUsuarioRequest struct {
	Nome  string `json:"nome"`
	Email string `json:"email"`
	Senha string `json:"senha"`
	Role  string `json:"role"`
}

type UpdateUsuarioRequest struct {
	Nome  string `json:"nome"`
	Email string `json:"email"`
	Senha string `json:"senha"` // Opcional na edição
	Role  string `json:"role"`
}

type AuthStatusResponse struct {
	NeedsSetup    bool     `json:"needs_setup"`
	Authenticated bool     `json:"authenticated"`
	Usuario       *Usuario `json:"usuario,omitempty"`
}

type AuthResponse struct {
	Token   string   `json:"token"`
	Usuario *Usuario `json:"usuario"`
}
