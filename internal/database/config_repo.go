package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"optimanager/internal/models"
)

// GetConfig retrieves the company/business institutional settings along with backup parameters.
func (d *DB) GetConfig() (*models.Config, error) {
	query := `SELECT id, fantasia, razao_social, cnpj, responsavel, fones, email_contato, endereco, cidade, 
	                 vt_valor_padrao, termo_garantia, termo_responsabilidade, obs_padrao, custom_presets, 
	                 logo_base64, cor_primaria, cor_secundaria,
	                 backup_auto_enabled, backup_frequencia, backup_dia_semana, backup_horario,
	                 backup_retencao_auto, backup_retencao_manual, backup_retencao_prerestore, backup_retencao_premigration,
	                 backup_ultimo_sucesso, backup_ultimo_erro, backup_ultimo_erro_em,
	                 backup_ultimo_aviso_externo, backup_ultimo_aviso_externo_em
	          FROM config 
	          WHERE id = 1`

	var c models.Config
	err := d.Reader.QueryRow(query).Scan(
		&c.ID, &c.Fantasia, &c.RazaoSocial, &c.CNPJ, &c.Responsavel, &c.Fones, &c.EmailContato, &c.Endereco, &c.Cidade,
		&c.VTValorPadrao, &c.TermoGarantia, &c.TermoResponsabilidade, &c.ObsPadrao, &c.CustomPresets,
		&c.LogoBase64, &c.CorPrimaria, &c.CorSecundaria,
		&c.BackupAutoEnabled, &c.BackupFrequencia, &c.BackupDiaSemana, &c.BackupHorario,
		&c.BackupRetencaoAuto, &c.BackupRetencaoManual, &c.BackupRetencaoPrerestore, &c.BackupRetencaoPremigration,
		&c.BackupUltimoSucesso, &c.BackupUltimoErro, &c.BackupUltimoErroEm,
		&c.BackupUltimoAvisoExterno, &c.BackupUltimoAvisoExternoEm,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("configuration record not found")
		}
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	return &c, nil
}

// UpdateConfig updates the institutional business settings.
func (d *DB) UpdateConfig(cfg *models.Config) error {
	query := `UPDATE config SET 
		fantasia = ?, razao_social = ?, cnpj = ?, responsavel = ?, 
		fones = ?, email_contato = ?, endereco = ?, cidade = ?, vt_valor_padrao = ?, 
		termo_garantia = ?, termo_responsabilidade = ?, obs_padrao = ?, custom_presets = ?, 
		logo_base64 = ?, cor_primaria = ?, cor_secundaria = ? 
		WHERE id = 1`

	res, err := d.Writer.Exec(query,
		strings.TrimSpace(cfg.Fantasia),
		strings.TrimSpace(cfg.RazaoSocial),
		strings.TrimSpace(cfg.CNPJ),
		strings.TrimSpace(cfg.Responsavel),
		strings.TrimSpace(cfg.Fones),
		strings.TrimSpace(cfg.EmailContato),
		strings.TrimSpace(cfg.Endereco),
		strings.TrimSpace(cfg.Cidade),
		cfg.VTValorPadrao,
		strings.TrimSpace(cfg.TermoGarantia),
		strings.TrimSpace(cfg.TermoResponsabilidade),
		strings.TrimSpace(cfg.ObsPadrao),
		strings.TrimSpace(cfg.CustomPresets),
		cfg.LogoBase64,
		strings.TrimSpace(cfg.CorPrimaria),
		strings.TrimSpace(cfg.CorSecundaria),
	)
	if err != nil {
		return fmt.Errorf("failed to update configuration: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("no configuration rows were updated")
	}
	return nil
}

// UpdateBackupConfig updates the scheduled backup routine parameters.
func (d *DB) UpdateBackupConfig(cfg *models.Config) error {
	query := `UPDATE config SET 
		backup_auto_enabled = ?,
		backup_frequencia = ?,
		backup_dia_semana = ?,
		backup_horario = ?,
		backup_retencao_auto = ?,
		backup_retencao_manual = ?,
		backup_retencao_prerestore = ?,
		backup_retencao_premigration = ?
		WHERE id = 1`

	res, err := d.Writer.Exec(query,
		cfg.BackupAutoEnabled,
		strings.TrimSpace(cfg.BackupFrequencia),
		cfg.BackupDiaSemana,
		strings.TrimSpace(cfg.BackupHorario),
		cfg.BackupRetencaoAuto,
		cfg.BackupRetencaoManual,
		cfg.BackupRetencaoPrerestore,
		cfg.BackupRetencaoPremigration,
	)
	if err != nil {
		return fmt.Errorf("failed to update backup configuration: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("no configuration rows were updated")
	}
	return nil
}

// SetBackupSuccess updates the last successful backup timestamp and clears previous local errors.
func (d *DB) SetBackupSuccess(timestamp string) error {
	query := `UPDATE config SET backup_ultimo_sucesso = ?, backup_ultimo_erro = '', backup_ultimo_erro_em = '' WHERE id = 1`
	_, err := d.Writer.Exec(query, timestamp)
	return err
}

// SetBackupError records the latest local backup failure message and timestamp.
func (d *DB) SetBackupError(errMsg, timestamp string) error {
	query := `UPDATE config SET backup_ultimo_erro = ?, backup_ultimo_erro_em = ? WHERE id = 1`
	_, err := d.Writer.Exec(query, errMsg, timestamp)
	return err
}

// SetBackupExternalWarning records the latest external sync warning message and timestamp.
func (d *DB) SetBackupExternalWarning(warningMsg, timestamp string) error {
	query := `UPDATE config SET backup_ultimo_aviso_externo = ?, backup_ultimo_aviso_externo_em = ? WHERE id = 1`
	_, err := d.Writer.Exec(query, warningMsg, timestamp)
	return err
}
