package database

import (
	"database/sql"
	"fmt"

	"optimanager/internal/models"
)

// GetDashboardData calcula todas as métricas em query agregada e retorna resumos para a tela inicial.
func (d *DB) GetDashboardData() (*models.DashboardData, error) {
	// 1. Métricas consolidadas em consulta única otimizada
	metricsQuery := `
		SELECT
			COALESCE(SUM(CASE WHEN data_emissao = DATE('now', 'localtime') AND status = 'Concluído' THEN valor_total ELSE 0 END), 0) AS faturamento_hoje,
			COALESCE(SUM(CASE WHEN status = 'Concluído' AND pagamento_status = 'Pago' THEN valor_total ELSE 0 END), 0) AS caixa_recebido,
			COALESCE(SUM(CASE WHEN status = 'Concluído' AND pagamento_status = 'Pendente' THEN valor_total ELSE 0 END), 0) AS pendente_receber,
			COALESCE(SUM(CASE WHEN status = 'Concluído' THEN valor_total ELSE 0 END), 0) AS faturamento_total,
			COALESCE(SUM(CASE WHEN status != 'Concluído' THEN 1 ELSE 0 END), 0) AS qtd_abertos,
			COALESCE(SUM(CASE WHEN status = 'Concluído' AND pagamento_status = 'Pendente' THEN 1 ELSE 0 END), 0) AS qtd_pendentes
		FROM ordens_servico;
	`

	var metricas models.DashboardMetricas
	err := d.Reader.QueryRow(metricsQuery).Scan(
		&metricas.FaturamentoHoje,
		&metricas.CaixaRecebido,
		&metricas.PendenteReceber,
		&metricas.FaturamentoTotal,
		&metricas.QtdAbertos,
		&metricas.QtdPendentes,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao calcular métricas do dashboard: %w", err)
	}

	// 2. Serviços em Aberto (limitado aos 10 mais recentes)
	abertosQuery := `
		SELECT os.id, os.data_emissao, os.status, os.cliente_id,
		       c.nome, c.doc, c.contato,
		       os.equipamento, os.servico, os.obs_tecnicas, os.valor_total,
		       os.pagamento_status, os.pagamento_forma, os.pago_em, os.created_at
		FROM ordens_servico os
		INNER JOIN clientes c ON c.id = os.cliente_id
		WHERE os.status != 'Concluído'
		ORDER BY os.id DESC
		LIMIT 10;
	`
	servicosAberto, err := d.fetchOrdens(abertosQuery)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar serviços em aberto: %w", err)
	}

	// 3. Cobranças Pendentes (concluídos aguardando pagamento, limitado aos 10 mais recentes)
	pendentesQuery := `
		SELECT os.id, os.data_emissao, os.status, os.cliente_id,
		       c.nome, c.doc, c.contato,
		       os.equipamento, os.servico, os.obs_tecnicas, os.valor_total,
		       os.pagamento_status, os.pagamento_forma, os.pago_em, os.created_at
		FROM ordens_servico os
		INNER JOIN clientes c ON c.id = os.cliente_id
		WHERE os.status = 'Concluído' AND os.pagamento_status = 'Pendente'
		ORDER BY os.id DESC
		LIMIT 10;
	`
	cobrancasPend, err := d.fetchOrdens(pendentesQuery)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar cobranças pendentes: %w", err)
	}

	// 4. Atalhos rápidos
	atalhos, err := d.ListAtalhos()
	if err != nil {
		return nil, fmt.Errorf("erro ao listar atalhos: %w", err)
	}

	return &models.DashboardData{
		Metricas:       metricas,
		ServicosAberto: servicosAberto,
		CobracasPend:   cobrancasPend,
		Atalhos:        atalhos,
	}, nil
}

func (d *DB) fetchOrdens(query string, args ...interface{}) ([]models.OrdemServico, error) {
	rows, err := d.Reader.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cnpj, fantasia string
	_ = d.Reader.QueryRow(`SELECT cnpj, fantasia FROM config WHERE id = 1`).Scan(&cnpj, &fantasia)

	res := make([]models.OrdemServico, 0, 10)
	for rows.Next() {
		var os models.OrdemServico
		var pagoEm sql.NullTime

		err := rows.Scan(
			&os.ID, &os.DataEmissao, &os.Status, &os.ClienteID,
			&os.ClienteNome, &os.ClienteDoc, &os.ClienteContato,
			&os.Equipamento, &os.Servico, &os.ObsTecnicas, &os.ValorTotal,
			&os.PagamentoStatus, &os.PagamentoForma, &pagoEm, &os.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		if pagoEm.Valid {
			os.PagoEm = &pagoEm.Time
		}
		os.WhatsAppURL = os.BuildWhatsAppURL(cnpj, fantasia)
		res = append(res, os)
	}
	return res, rows.Err()
}

