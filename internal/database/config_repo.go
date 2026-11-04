package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"optimanager/internal/models"
)

// GetConfig retorna a configuração institucional da empresa/MEI.
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
			return nil, errors.New("configuração não encontrada")
		}
		return nil, fmt.Errorf("erro ao carregar configurações: %w", err)
	}
	return &c, nil
}

// UpdateConfig atualiza os dados cadastrais da empresa/MEI.
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
		return fmt.Errorf("erro ao atualizar configurações: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("nenhuma linha de configuração foi atualizada")
	}
	return nil
}
