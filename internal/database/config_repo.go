package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"optimanager/internal/models"
)

// GetConfig retrieves the company/business institutional settings.
func (d *DB) GetConfig() (*models.Config, error) {
	query := `SELECT id, fantasia, razao_social, cnpj, responsavel, fones, email_contato, endereco, cidade, 
	                 vt_valor_padrao, termo_garantia, logo_base64, cor_primaria, cor_secundaria 
	          FROM config 
	          WHERE id = 1`

	var c models.Config
	err := d.Reader.QueryRow(query).Scan(
		&c.ID, &c.Fantasia, &c.RazaoSocial, &c.CNPJ, &c.Responsavel, &c.Fones, &c.EmailContato, &c.Endereco, &c.Cidade,
		&c.VTValorPadrao, &c.TermoGarantia, &c.LogoBase64, &c.CorPrimaria, &c.CorSecundaria,
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
		fones = ?, email_contato = ?, endereco = ?, cidade = ?, vt_valor_padrao = ?, termo_garantia = ?, 
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
