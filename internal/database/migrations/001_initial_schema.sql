-- Schema inicial do OptiManager

CREATE TABLE IF NOT EXISTS config (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    fantasia TEXT NOT NULL DEFAULT 'Opti Tecnologia',
    razao_social TEXT NOT NULL DEFAULT 'João Vítor Streb Figueiró',
    cnpj TEXT NOT NULL DEFAULT '38.136.905/0001-66',
    responsavel TEXT NOT NULL DEFAULT 'João Figueiró',
    fones TEXT NOT NULL DEFAULT '(55) 99127-2332 | (55) 99674-5762',
    email_contato TEXT NOT NULL DEFAULT '',
    endereco TEXT NOT NULL DEFAULT '',
    cidade TEXT NOT NULL DEFAULT 'Santa Maria - RS',
    vt_valor_padrao REAL NOT NULL DEFAULT 80.00,
    termo_garantia TEXT NOT NULL DEFAULT 'Garantia de 90 dias para serviços prestados e peças substituídas.',
    logo_base64 TEXT NOT NULL DEFAULT '',
    cor_primaria TEXT NOT NULL DEFAULT '#0B5694',
    cor_secundaria TEXT NOT NULL DEFAULT '#083f6e'
);

-- Insere configuração padrão caso não exista
INSERT OR IGNORE INTO config (id, fantasia, razao_social, cnpj, responsavel, fones, email_contato, endereco, cidade, vt_valor_padrao, termo_garantia, logo_base64, cor_primaria, cor_secundaria)
VALUES (
    1,
    'Opti Tecnologia',
    'João Vítor Streb Figueiró',
    '38.136.905/0001-66',
    'João Figueiró',
    '(55) 99127-2332 | (55) 99674-5762',
    '',
    '',
    'Santa Maria - RS',
    80.00,
    'Garantia legal de 90 dias para serviços executados e peças aplicadas.',
    '',
    '#0B5694',
    '#083f6e'
);

CREATE TABLE IF NOT EXISTS clientes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    nome TEXT NOT NULL UNIQUE COLLATE NOCASE,
    doc TEXT NOT NULL DEFAULT '',
    contato TEXT NOT NULL DEFAULT '',
    desde_ano INTEGER NOT NULL DEFAULT (strftime('%Y', 'now')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_clientes_nome ON clientes(nome);

CREATE TABLE IF NOT EXISTS ordens_servico (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    data_emissao DATE NOT NULL,
    status TEXT NOT NULL DEFAULT 'Avaliação' CHECK (status IN ('Avaliação', 'Orçamento', 'Em Serviço', 'Concluído')),
    cliente_id INTEGER NOT NULL,
    equipamento TEXT NOT NULL DEFAULT 'N/A',
    servico TEXT NOT NULL,
    obs_tecnicas TEXT NOT NULL DEFAULT '',
    valor_total REAL NOT NULL DEFAULT 0.00,
    pagamento_status TEXT NOT NULL DEFAULT 'Pendente' CHECK (pagamento_status IN ('Pendente', 'Pago')),
    pagamento_forma TEXT NOT NULL DEFAULT 'Pix' CHECK (pagamento_forma IN ('Pix', 'Dinheiro', 'Cartão', 'Boleto')),
    pago_em DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (cliente_id) REFERENCES clientes(id) ON DELETE RESTRICT
);

-- Inicializa o ID da primeira OS para 1000 (sqlite_sequence com 999)
INSERT OR IGNORE INTO sqlite_sequence (name, seq) VALUES ('ordens_servico', 999);

CREATE INDEX IF NOT EXISTS idx_os_data_emissao ON ordens_servico(data_emissao);
CREATE INDEX IF NOT EXISTS idx_os_status ON ordens_servico(status);
CREATE INDEX IF NOT EXISTS idx_os_pagamento_status ON ordens_servico(pagamento_status);
CREATE INDEX IF NOT EXISTS idx_os_cliente_id ON ordens_servico(cliente_id);

CREATE TABLE IF NOT EXISTS atalhos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    titulo TEXT NOT NULL,
    descricao TEXT NOT NULL DEFAULT '',
    url TEXT NOT NULL,
    ordem INTEGER NOT NULL DEFAULT 0
);

-- Inserção de atalhos padrão do MEI se a tabela estiver vazia
INSERT OR IGNORE INTO atalhos (id, titulo, descricao, url, ordem) VALUES
(1, 'PGMEI - Guia DAS', 'Gerador de DAS do MEI', 'http://www8.receita.fazenda.gov.br/SimplesNacional/Aplicacoes/ATSPO/pgmei.app/Identificacao', 1),
(2, 'ISSNet Santa Maria', 'Nota Fiscal Eletrônica Municipal', 'https://www.issnetonline.com.br/santamaria/online/login/login.aspx', 2),
(3, 'Portal do Empreendedor', 'Serviços e Declaração Anual (DASN)', 'https://www.gov.br/empresas-e-negocios/pt-br/empreendedor/servicos-para-mei', 3),
(4, 'Consulta Cartão CNPJ', 'Comprovante de Inscrição Receita', 'https://solucoes.receita.fazenda.gov.br/Servicos/cnpjreva/Cnpjreva_Solicitacao.asp', 4);

-- Tabela de Usuários para Autenticação e Controle de Acesso
CREATE TABLE IF NOT EXISTS usuarios (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    nome TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE COLLATE NOCASE,
    senha_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'admin' CHECK (role IN ('admin', 'operador')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_usuarios_email ON usuarios(email);
