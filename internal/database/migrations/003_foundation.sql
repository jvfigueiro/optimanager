-- Migration 003: Foundation (Centavos, Optimistic Concurrency, Inactive Users, Permissions, Audit Entity Columns)

-- Ensure baseline tables exist for minimal v1 test databases
CREATE TABLE IF NOT EXISTS clientes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    nome TEXT NOT NULL UNIQUE COLLATE NOCASE,
    doc TEXT NOT NULL DEFAULT '',
    contato TEXT NOT NULL DEFAULT '',
    desde_ano INTEGER NOT NULL DEFAULT (strftime('%Y', 'now')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS ordens_servico (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    data_emissao DATE NOT NULL,
    hora_emissao TEXT NOT NULL DEFAULT '',
    data_entrega TEXT NOT NULL DEFAULT '',
    hora_entrega TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'Avaliação',
    cliente_id INTEGER NOT NULL,
    equipamento TEXT NOT NULL DEFAULT 'N/A',
    servico TEXT NOT NULL,
    obs_tecnicas TEXT NOT NULL DEFAULT '',
    valor_total REAL NOT NULL DEFAULT 0.00,
    pagamento_status TEXT NOT NULL DEFAULT 'Pendente' CHECK (pagamento_status IN ('Pendente', 'Pago')),
    pagamento_forma TEXT NOT NULL DEFAULT 'Pix',
    pago_em DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (cliente_id) REFERENCES clientes(id) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS atalhos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    titulo TEXT NOT NULL,
    descricao TEXT NOT NULL DEFAULT '',
    url TEXT NOT NULL,
    ordem INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS usuarios (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    nome TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE COLLATE NOCASE,
    senha_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'admin' CHECK (role IN ('admin', 'operador')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 1. Extend config with centavos, optimistic locking version, and active modules
ALTER TABLE config ADD COLUMN vt_valor_padrao_centavos INTEGER NOT NULL DEFAULT 8000;
ALTER TABLE config ADD COLUMN versao INTEGER NOT NULL DEFAULT 1;
ALTER TABLE config ADD COLUMN modulos_ativos TEXT NOT NULL DEFAULT '{"estoque":true,"compras":true,"vendas":false,"financeiro":true}';

UPDATE config
SET vt_valor_padrao_centavos = CAST(ROUND(COALESCE(vt_valor_padrao, 80.0) * 100.0) AS INTEGER),
    cnpj = CASE WHEN TRIM(cnpj) = '00.000.000/0001-00' THEN '' ELSE TRIM(cnpj) END;

-- 2. Add optimistic locking version to clientes
ALTER TABLE clientes ADD COLUMN versao INTEGER NOT NULL DEFAULT 1;

-- 3. Add centavos and optimistic locking version to ordens_servico
ALTER TABLE ordens_servico ADD COLUMN valor_total_centavos INTEGER NOT NULL DEFAULT 0;
ALTER TABLE ordens_servico ADD COLUMN versao INTEGER NOT NULL DEFAULT 1;

UPDATE ordens_servico
SET valor_total_centavos = CAST(ROUND(COALESCE(valor_total, 0.0) * 100.0) AS INTEGER);

-- 4. Rebuild usuarios to support optional email, inactive legacy technicians, profiles, and version
CREATE TABLE usuarios_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    nome TEXT NOT NULL,
    email TEXT DEFAULT NULL COLLATE NOCASE,
    senha_hash TEXT NOT NULL DEFAULT '',
    role TEXT NOT NULL DEFAULT 'admin',
    perfil TEXT NOT NULL DEFAULT 'administrador' CHECK (perfil IN ('administrador', 'atendente', 'tecnico')),
    ativo INTEGER NOT NULL DEFAULT 1 CHECK (ativo IN (0, 1)),
    versao INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO usuarios_new (id, nome, email, senha_hash, role, perfil, ativo, versao, created_at)
SELECT
    id,
    nome,
    CASE WHEN TRIM(COALESCE(email, '')) = '' THEN NULL ELSE TRIM(LOWER(email)) END,
    COALESCE(senha_hash, ''),
    CASE WHEN role = 'operador' THEN 'operador' ELSE 'admin' END,
    CASE WHEN role = 'operador' THEN 'atendente' ELSE 'administrador' END,
    1,
    1,
    created_at
FROM usuarios;

DROP TABLE usuarios;
ALTER TABLE usuarios_new RENAME TO usuarios;

CREATE UNIQUE INDEX IF NOT EXISTS idx_usuarios_email_unique
ON usuarios(email COLLATE NOCASE)
WHERE email IS NOT NULL AND email != '';

-- 5. Create permission tables and seed default profile permissions
CREATE TABLE IF NOT EXISTS perfil_permissoes (
    perfil TEXT NOT NULL CHECK (perfil IN ('administrador', 'atendente', 'tecnico')),
    permissao TEXT NOT NULL,
    PRIMARY KEY (perfil, permissao)
);

CREATE TABLE IF NOT EXISTS usuario_permissoes (
    usuario_id INTEGER NOT NULL,
    permissao TEXT NOT NULL,
    concedida INTEGER NOT NULL DEFAULT 1 CHECK (concedida IN (0, 1)),
    PRIMARY KEY (usuario_id, permissao),
    FOREIGN KEY (usuario_id) REFERENCES usuarios(id) ON DELETE CASCADE
);

-- Seed administrator permissions
INSERT OR IGNORE INTO perfil_permissoes (perfil, permissao) VALUES
('administrador', 'ver_valores'),
('administrador', 'ver_dados_pessoais'),
('administrador', 'dar_desconto'),
('administrador', 'cancelar_os'),
('administrador', 'reabrir_os'),
('administrador', 'excluir'),
('administrador', 'ver_financeiro'),
('administrador', 'gerenciar_catalogo'),
('administrador', 'gerenciar_clientes'),
('administrador', 'exportar'),
('administrador', 'editar_numero_os'),
('administrador', 'gerenciar_usuarios'),
('administrador', 'gerenciar_config');

-- Seed atendente permissions (matches legacy operador capabilities)
INSERT OR IGNORE INTO perfil_permissoes (perfil, permissao) VALUES
('atendente', 'ver_valores'),
('atendente', 'ver_dados_pessoais'),
('atendente', 'dar_desconto'),
('atendente', 'cancelar_os'),
('atendente', 'reabrir_os'),
('atendente', 'excluir'),
('atendente', 'ver_financeiro'),
('atendente', 'gerenciar_catalogo'),
('atendente', 'gerenciar_clientes'),
('atendente', 'exportar');

-- Seed tecnico permissions
INSERT OR IGNORE INTO perfil_permissoes (perfil, permissao) VALUES
('tecnico', 'ver_dados_pessoais'),
('tecnico', 'gerenciar_catalogo'),
('tecnico', 'gerenciar_clientes');

-- 6. Extend audit_log with entity metadata columns
ALTER TABLE audit_log ADD COLUMN entidade TEXT NOT NULL DEFAULT '';
ALTER TABLE audit_log ADD COLUMN entidade_id TEXT NOT NULL DEFAULT '';
