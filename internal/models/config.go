package models

type Config struct {
	ID            int     `json:"id"`
	Fantasia      string  `json:"fantasia"`
	RazaoSocial   string  `json:"razao_social"`
	CNPJ          string  `json:"cnpj"`
	Responsavel   string  `json:"responsavel"`
	Fones         string  `json:"fones"`
	EmailContato  string  `json:"email_contato"`
	Endereco      string  `json:"endereco"`
	Cidade        string  `json:"cidade"`
	VTValorPadrao float64 `json:"vt_valor_padrao"`
	TermoGarantia         string  `json:"termo_garantia"`
	TermoResponsabilidade string  `json:"termo_responsabilidade"`
	ObsPadrao             string  `json:"obs_padrao"`
	CustomPresets         string  `json:"custom_presets"`
	LogoBase64            string  `json:"logo_base64"`
	CorPrimaria   string  `json:"cor_primaria"`
	CorSecundaria string  `json:"cor_secundaria"`
}
