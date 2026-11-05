package database

import (
	"fmt"

	"optimanager/internal/models"
)

// InsertAuditLog records an audit event in the database.
func (d *DB) InsertAuditLog(usuarioID int64, usuarioNome, ip, acao, detalhe string) error {
	query := `INSERT INTO audit_log (usuario_id, usuario_nome, ip, acao, detalhe) VALUES (?, ?, ?, ?, ?)`
	_, err := d.Writer.Exec(query, usuarioID, usuarioNome, ip, acao, detalhe)
	if err != nil {
		return fmt.Errorf("failed to insert audit log: %w", err)
	}
	return nil
}

// ListAuditLogs retrieves the most recent audit logs ordered by creation timestamp.
func (d *DB) ListAuditLogs(limit int) ([]models.AuditLog, error) {
	if limit <= 0 {
		limit = 100
	}
	query := `SELECT id, usuario_id, usuario_nome, ip, acao, detalhe, created_at FROM audit_log ORDER BY created_at DESC, id DESC LIMIT ?`
	rows, err := d.Reader.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		var a models.AuditLog
		if err := rows.Scan(&a.ID, &a.UsuarioID, &a.UsuarioNome, &a.IP, &a.Acao, &a.Detalhe, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}
		logs = append(logs, a)
	}
	return logs, rows.Err()
}
