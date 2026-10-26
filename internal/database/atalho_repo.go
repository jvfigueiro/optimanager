package database

import (
	"errors"
	"fmt"
	"strings"

	"optimanager/internal/models"
)

// ListAtalhos retorna os links e atalhos rápidos ordenados.
func (d *DB) ListAtalhos() ([]models.Atalho, error) {
	query := `SELECT id, titulo, descricao, url, ordem FROM atalhos ORDER BY ordem ASC, id ASC`
	rows, err := d.Reader.Query(query)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar atalhos: %w", err)
	}
	defer rows.Close()

	atalhos := make([]models.Atalho, 0, 8)
	for rows.Next() {
		var a models.Atalho
		if err := rows.Scan(&a.ID, &a.Titulo, &a.Descricao, &a.Url, &a.Ordem); err != nil {
			return nil, fmt.Errorf("erro ao ler atalho: %w", err)
		}
		atalhos = append(atalhos, a)
	}
	return atalhos, rows.Err()
}

// CreateAtalho cadastra um novo atalho rápido no sistema.
func (d *DB) CreateAtalho(a *models.Atalho) error {
	a.Titulo = strings.TrimSpace(a.Titulo)
	a.Url = strings.TrimSpace(a.Url)
	if a.Titulo == "" || a.Url == "" {
		return errors.New("título e URL do atalho são obrigatórios")
	}

	query := `INSERT INTO atalhos (titulo, descricao, url, ordem) VALUES (?, ?, ?, ?)`
	res, err := d.Writer.Exec(query, a.Titulo, strings.TrimSpace(a.Descricao), a.Url, a.Ordem)
	if err != nil {
		return fmt.Errorf("erro ao criar atalho: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	a.ID = id
	return nil
}

// DeleteAtalho remove um atalho pelo ID.
func (d *DB) DeleteAtalho(id int64) error {
	res, err := d.Writer.Exec(`DELETE FROM atalhos WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("erro ao deletar atalho: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("atalho não encontrado")
	}
	return nil
}

