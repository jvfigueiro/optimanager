package models

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type OrdemServico struct {
	ID              int64      `json:"id"`
	DataEmissao     string     `json:"data_emissao"`
	HoraEmissao     string     `json:"hora_emissao"`
	DataEntrega     string     `json:"data_entrega"`
	HoraEntrega     string     `json:"hora_entrega"`
	Status          string     `json:"status"`
	ClienteID       int64      `json:"cliente_id"`
	ClienteNome     string     `json:"cliente_nome,omitempty"`
	ClienteDoc      string     `json:"cliente_doc,omitempty"`
	ClienteContato  string     `json:"cliente_contato,omitempty"`
	Equipamento     string     `json:"equipamento"`
	Servico         string     `json:"servico"`
	ObsTecnicas     string     `json:"obs_tecnicas"`
	ValorTotal      float64    `json:"valor_total"`
	PagamentoStatus string     `json:"pagamento_status"` // 'Pendente', 'Pago'
	PagamentoForma  string     `json:"pagamento_forma"`
	PagoEm          *time.Time `json:"pago_em,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	WhatsAppURL     string     `json:"whatsapp_url,omitempty"`
}

// CreateOrdemRequest defines payload to create a new service order.
type CreateOrdemRequest struct {
	DataEmissao     string  `json:"data_emissao"`
	HoraEmissao     string  `json:"hora_emissao"`
	DataEntrega     string  `json:"data_entrega"`
	HoraEntrega     string  `json:"hora_entrega"`
	Status          string  `json:"status"`
	ClienteID       int64   `json:"cliente_id"`
	ClienteNome     string  `json:"cliente_nome"`
	ClienteDoc      string  `json:"cliente_doc"`
	ClienteContato  string  `json:"cliente_contato"`
	Equipamento     string  `json:"equipamento"`
	Servico         string  `json:"servico"`
	ObsTecnicas     string  `json:"obs_tecnicas"`
	ValorTotal      float64 `json:"valor_total"`
	PagamentoStatus string  `json:"pagamento_status"`
	PagamentoForma  string  `json:"pagamento_forma"`
	IsVT            bool    `json:"is_vt"` // When true, triggers On-site Technical Visit shortcut
}

// UpdateOrdemRequest defines payload to update an existing service order.
type UpdateOrdemRequest struct {
	DataEmissao     string  `json:"data_emissao"`
	HoraEmissao     string  `json:"hora_emissao"`
	DataEntrega     string  `json:"data_entrega"`
	HoraEntrega     string  `json:"hora_entrega"`
	Status          string  `json:"status"`
	ClienteID       int64   `json:"cliente_id"`
	ClienteNome     string  `json:"cliente_nome"`
	ClienteDoc      string  `json:"cliente_doc"`
	ClienteContato  string  `json:"cliente_contato"`
	Equipamento     string  `json:"equipamento"`
	Servico         string  `json:"servico"`
	ObsTecnicas     string  `json:"obs_tecnicas"`
	ValorTotal      float64 `json:"valor_total"`
	PagamentoStatus string  `json:"pagamento_status"`
	PagamentoForma  string  `json:"pagamento_forma"`
}

// BaixaPagamentoRequest defines payload for quick payment settlement.
type BaixaPagamentoRequest struct {
	Forma  string `json:"forma"`  // 'Pix', 'Dinheiro', 'Cartão', 'Boleto'
	Status string `json:"status"` // 'Pago'
}

// BatchOrdensRequest defines payload for bulk operations on service orders.
type BatchOrdensRequest struct {
	Action         string  `json:"action"` // "baixa", "status", "duplicar", "delete"
	IDs            []int64 `json:"ids"`
	Status         string  `json:"status,omitempty"`
	PagamentoForma string  `json:"pagamento_forma,omitempty"`
}

var nonDigitsRegex = regexp.MustCompile(`\D+`)

// BuildWhatsAppURL builds a wa.me URL with a formatted summary of the order.
func (os *OrdemServico) BuildWhatsAppURL(cnpj, empresa string) string {
	rawPhone := strings.TrimSpace(os.ClienteContato)
	cleanPhone := nonDigitsRegex.ReplaceAllString(rawPhone, "")
	if cleanPhone == "" {
		return ""
	}

	// Prepend Brazil country code (55) if only area code + number provided
	if len(cleanPhone) == 10 || len(cleanPhone) == 11 {
		cleanPhone = "55" + cleanPhone
	}

	empresaNome := "Minha Empresa Tech"
	if empresa != "" {
		empresaNome = empresa
	}

	pixKey := "00.000.000/0001-00"
	if cnpj != "" {
		pixKey = cnpj
	}

	msg := fmt.Sprintf(
		"Olá, %s!\n\nSegue o resumo da sua Ordem de Serviço #%d:\n• Situação: %s\n• Equipamento: %s\n• Serviço: %s\n• Valor Total: R$ %.2f\n• Pagamento: %s (%s)\n\nChave Pix (CNPJ): %s\n\n%s à sua disposição!",
		os.ClienteNome,
		os.ID,
		os.Status,
		os.Equipamento,
		os.Servico,
		os.ValorTotal,
		os.PagamentoStatus,
		os.PagamentoForma,
		pixKey,
		empresaNome,
	)

	return fmt.Sprintf("https://wa.me/%s?text=%s", cleanPhone, url.QueryEscape(msg))
}

