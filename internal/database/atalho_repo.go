package database

import (
	"errors"
	"fmt"
	"strings"

	"optimanager/internal/models"
)

// ListAtalhos returns quick portal links ordered by display sequence.
func (d *DB) ListAtalhos() ([]models.Atalho, error) {
	query := `SELECT id, titulo, descricao, url, ordem FROM atalhos ORDER BY ordem ASC, id ASC`
	rows, err := d.Reader.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query shortcuts: %w", err)
	}
	defer rows.Close()

	atalhos := make([]models.Atalho, 0, 8)
	for rows.Next() {
		var a models.Atalho
		if err := rows.Scan(&a.ID, &a.Titulo, &a.Descricao, &a.Url, &a.Ordem); err != nil {
			return nil, fmt.Errorf("failed to scan shortcut: %w", err)
		}
		atalhos = append(atalhos, a)
	}
	return atalhos, rows.Err()
}

// CreateAtalho registers a new portal shortcut.
func (d *DB) CreateAtalho(a *models.Atalho) error {
	a.Titulo = strings.TrimSpace(a.Titulo)
	a.Url = strings.TrimSpace(a.Url)
	if a.Titulo == "" || a.Url == "" {
		return errors.New("shortcut title and URL are required")
	}

	query := `INSERT INTO atalhos (titulo, descricao, url, ordem) VALUES (?, ?, ?, ?)`
	res, err := d.Writer.Exec(query, a.Titulo, strings.TrimSpace(a.Descricao), a.Url, a.Ordem)
	if err != nil {
		return fmt.Errorf("failed to create shortcut: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	a.ID = id
	return nil
}

// UpdateAtalho updates an existing portal shortcut.
func (d *DB) UpdateAtalho(a *models.Atalho) error {
	a.Titulo = strings.TrimSpace(a.Titulo)
	a.Url = strings.TrimSpace(a.Url)
	if a.Titulo == "" || a.Url == "" {
		return errors.New("shortcut title and URL are required")
	}

	query := `UPDATE atalhos SET titulo = ?, descricao = ?, url = ?, ordem = ? WHERE id = ?`
	res, err := d.Writer.Exec(query, a.Titulo, strings.TrimSpace(a.Descricao), a.Url, a.Ordem, a.ID)
	if err != nil {
		return fmt.Errorf("failed to update shortcut: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("shortcut not found")
	}
	return nil
}

// DeleteAtalho removes a portal shortcut by ID.
func (d *DB) DeleteAtalho(id int64) error {
	res, err := d.Writer.Exec(`DELETE FROM atalhos WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete shortcut: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("shortcut not found")
	}
	return nil
}
