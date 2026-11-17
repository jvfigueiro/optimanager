package models

type DashboardMetricas struct {
	FaturamentoHoje          float64 `json:"faturamento_hoje"`
	FaturamentoHojeCentavos  int64   `json:"faturamento_hoje_centavos"`
	CaixaRecebido            float64 `json:"caixa_recebido"`
	CaixaRecebidoCentavos    int64   `json:"caixa_recebido_centavos"`
	PendenteReceber          float64 `json:"pendente_receber"`
	PendenteReceberCentavos  int64   `json:"pendente_receber_centavos"`
	FaturamentoTotal         float64 `json:"faturamento_total"`
	FaturamentoTotalCentavos int64   `json:"faturamento_total_centavos"`
	QtdAbertos               int     `json:"qtd_abertos"`
	QtdPendentes             int     `json:"qtd_pendentes"`
}

type DashboardData struct {
	Metricas       DashboardMetricas `json:"metricas"`
	ServicosAberto []OrdemServico    `json:"servicos_aberto"`
	CobracasPend   []OrdemServico    `json:"cobrancas_pendentes"`
	Atalhos        []Atalho          `json:"atalhos"`
}

// RedactValues removes all monetary figures when the user lacks the "ver_valores" permission.
func (d *DashboardData) RedactValues() {
	if d == nil {
		return
	}
	d.Metricas.FaturamentoHoje = 0
	d.Metricas.FaturamentoHojeCentavos = 0
	d.Metricas.CaixaRecebido = 0
	d.Metricas.CaixaRecebidoCentavos = 0
	d.Metricas.PendenteReceber = 0
	d.Metricas.PendenteReceberCentavos = 0
	d.Metricas.FaturamentoTotal = 0
	d.Metricas.FaturamentoTotalCentavos = 0

	for i := range d.ServicosAberto {
		d.ServicosAberto[i].RedactValues()
	}
	for i := range d.CobracasPend {
		d.CobracasPend[i].RedactValues()
	}
}
