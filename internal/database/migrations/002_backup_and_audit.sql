-- Migration 002: Add backup configuration columns and audit_log table

-- Add backup columns to config table
ALTER TABLE config ADD COLUMN backup_auto_enabled INTEGER NOT NULL DEFAULT 0;
ALTER TABLE config ADD COLUMN backup_frequencia TEXT NOT NULL DEFAULT 'diario';
ALTER TABLE config ADD COLUMN backup_dia_semana INTEGER NOT NULL DEFAULT 0;
ALTER TABLE config ADD COLUMN backup_horario TEXT NOT NULL DEFAULT '02:00';
ALTER TABLE config ADD COLUMN backup_retencao_auto INTEGER NOT NULL DEFAULT 7;
ALTER TABLE config ADD COLUMN backup_retencao_manual INTEGER NOT NULL DEFAULT 5;
ALTER TABLE config ADD COLUMN backup_retencao_prerestore INTEGER NOT NULL DEFAULT 3;
ALTER TABLE config ADD COLUMN backup_retencao_premigration INTEGER NOT NULL DEFAULT 3;
ALTER TABLE config ADD COLUMN backup_ultimo_sucesso TEXT NOT NULL DEFAULT '';
ALTER TABLE config ADD COLUMN backup_ultimo_erro TEXT NOT NULL DEFAULT '';
ALTER TABLE config ADD COLUMN backup_ultimo_erro_em TEXT NOT NULL DEFAULT '';
ALTER TABLE config ADD COLUMN backup_ultimo_aviso_externo TEXT NOT NULL DEFAULT '';
ALTER TABLE config ADD COLUMN backup_ultimo_aviso_externo_em TEXT NOT NULL DEFAULT '';

-- Create audit_log table
CREATE TABLE IF NOT EXISTS audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    usuario_id INTEGER NOT NULL,
    usuario_nome TEXT NOT NULL,
    ip TEXT NOT NULL,
    acao TEXT NOT NULL,
    detalhe TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_log_created_at ON audit_log (created_at DESC, id DESC);
