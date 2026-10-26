package models

type Config struct {
	ID            int     `json:"id"`
	Fantasia      string  `json:"fantasia"`
	RazaoSocial   string  `json:"razao_social"`
	CNPJ          string  `json:"cnpj"`
	Responsavel   string  `json:"responsavel"`
	Fones         string  `json:"fones"`
	Cidade        string  `json:"cidade"`
	VTValorPadrao float64 `json:"vt_valor_padrao"`
	TermoGarantia string  `json:"termo_garantia"`
	LogoBase64    string  `json:"logo_base64"`
	CorPrimaria   string  `json:"cor_primaria"`
	CorSecundaria string  `json:"cor_secundaria"`
}

