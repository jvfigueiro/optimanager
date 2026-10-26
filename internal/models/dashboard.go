package models

type DashboardMetricas struct {
	FaturamentoHoje  float64 `json:"faturamento_hoje"`
	CaixaRecebido    float64 `json:"caixa_recebido"`
	PendenteReceber  float64 `json:"pendente_receber"`
	FaturamentoTotal float64 `json:"faturamento_total"`
	QtdAbertos       int     `json:"qtd_abertos"`
	QtdPendentes     int     `json:"qtd_pendentes"`
}

type DashboardData struct {
	Metricas       DashboardMetricas `json:"metricas"`
	ServicosAberto []OrdemServico    `json:"servicos_aberto"`
	CobracasPend   []OrdemServico    `json:"cobrancas_pendentes"`
	Atalhos        []Atalho          `json:"atalhos"`
}

