package models

import "time"

// Permission constants used across handlers and serializers.
const (
	PermVerValores        = "ver_valores"
	PermVerDadosPessoais  = "ver_dados_pessoais"
	PermDarDesconto       = "dar_desconto"
	PermCancelarOS        = "cancelar_os"
	PermReabrirOS         = "reabrir_os"
	PermExcluir           = "excluir"
	PermVerFinanceiro     = "ver_financeiro"
	PermGerenciarCatalogo = "gerenciar_catalogo"
	PermGerenciarClientes = "gerenciar_clientes"
	PermExportar          = "exportar"
	PermEditarNumeroOS    = "editar_numero_os"
	PermGerenciarUsuarios = "gerenciar_usuarios"
	PermGerenciarConfig   = "gerenciar_config"
)

type Usuario struct {
	ID         int64     `json:"id"`
	Nome       string    `json:"nome"`
	Email      string    `json:"email"`
	SenhaHash  string    `json:"-"`
	Role       string    `json:"role"`   // 'admin' | 'operador' (kept in sync with Perfil)
	Perfil     string    `json:"perfil"` // 'administrador' | 'atendente' | 'tecnico'
	Ativo      bool      `json:"ativo"`
	Versao     int64     `json:"versao"`
	Permissoes []string  `json:"permissoes"`
	CreatedAt  time.Time `json:"created_at"`
}

// HasPermission returns true if the user is active and holds the requested permission
// (or is an administrator).
func (u *Usuario) HasPermission(perm string) bool {
	if u == nil || !u.Ativo {
		return false
	}
	if u.Role == "admin" || u.Perfil == "administrador" {
		return true
	}
	for _, p := range u.Permissoes {
		if p == perm {
			return true
		}
	}
	return false
}

type SetupAdminRequest struct {
	// Administrator credentials
	Nome  string `json:"nome"`
	Email string `json:"email"`
	Senha string `json:"senha"`

	// Company details configured during initial wizard
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
	Nome   string `json:"nome"`
	Email  string `json:"email"`
	Senha  string `json:"senha"`
	Role   string `json:"role"`
	Perfil string `json:"perfil,omitempty"`
	Ativo  *bool  `json:"ativo,omitempty"`
}

type UpdateUsuarioRequest struct {
	Nome   string `json:"nome"`
	Email  string `json:"email"`
	Senha  string `json:"senha"` // Optional on profile update
	Role   string `json:"role"`
	Perfil string `json:"perfil,omitempty"`
	Ativo  *bool  `json:"ativo,omitempty"`
	Versao int64  `json:"versao,omitempty"`
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
