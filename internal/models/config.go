package models

type Config struct {
	ID                    int     `json:"id"`
	Fantasia              string  `json:"fantasia"`
	RazaoSocial           string  `json:"razao_social"`
	CNPJ                  string  `json:"cnpj"`
	Responsavel           string  `json:"responsavel"`
	Fones                 string  `json:"fones"`
	EmailContato          string  `json:"email_contato"`
	Endereco              string  `json:"endereco"`
	Cidade                string  `json:"cidade"`
	VTValorPadrao         float64 `json:"vt_valor_padrao"`
	TermoGarantia         string  `json:"termo_garantia"`
	TermoResponsabilidade string  `json:"termo_responsabilidade"`
	ObsPadrao             string  `json:"obs_padrao"`
	CustomPresets         string  `json:"custom_presets"`
	LogoBase64            string  `json:"logo_base64"`
	CorPrimaria           string  `json:"cor_primaria"`
	CorSecundaria         string  `json:"cor_secundaria"`

	// Backup configuration fields
	BackupAutoEnabled          int    `json:"backup_auto_enabled"`
	BackupFrequencia           string `json:"backup_frequencia"`
	BackupDiaSemana            int    `json:"backup_dia_semana"`
	BackupHorario              string `json:"backup_horario"`
	BackupRetencaoAuto         int    `json:"backup_retencao_auto"`
	BackupRetencaoManual       int    `json:"backup_retencao_manual"`
	BackupRetencaoPrerestore   int    `json:"backup_retencao_prerestore"`
	BackupRetencaoPremigration int    `json:"backup_retencao_premigration"`
	BackupUltimoSucesso        string `json:"backup_ultimo_sucesso"`
	BackupUltimoErro           string `json:"backup_ultimo_erro"`
	BackupUltimoErroEm         string `json:"backup_ultimo_erro_em"`
	BackupUltimoAvisoExterno   string `json:"backup_ultimo_aviso_externo"`
	BackupUltimoAvisoExternoEm string `json:"backup_ultimo_aviso_externo_em"`
}
