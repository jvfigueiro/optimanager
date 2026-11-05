<div align="center">

# 💼 OptiManager

**Sistema de Gestão Empresarial e Ordens de Serviço Auto-Hospedado (On-Premises)**

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Vue 3](https://img.shields.io/badge/Vue-3.5+-4FC08D?style=flat&logo=vuedotjs)](https://vuejs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.7+-3178C6?style=flat&logo=typescript)](https://www.typescriptlang.org/)
[![SQLite](https://img.shields.io/badge/SQLite-WAL%20Mode-003B57?style=flat&logo=sqlite)](https://www.sqlite.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![PWA](https://img.shields.io/badge/PWA-Installable-5A0FC8?style=flat&logo=pwa)](https://developer.mozilla.org/en-US/docs/Web/Progressive_web_apps)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

*Uma aplicação web rápida, leve e auto-suficiente projetada para micro e pequenos empreendedores, técnicos e prestadores de serviços gerenciarem atendimentos, clientes, relatórios técnicos e faturamento com controle total dos seus próprios dados.*

</div>

---

## 🌟 Principais Recursos

- 🚀 **Assistente de Primeiro Uso Guiado:** Configuração inicial rápida em 6 passos (boas-vindas, termos de uso, opções de nova instalação ou restauração de backup, criação do administrador e dados da empresa).
- 📋 **Gestão Completa de Ordens de Serviço (OS):** Criação, edição, filtragem por status (Orçamento, Aberto, Em Andamento, Concluído, Entregue, Cancelado), responsável, prazos e controle de valores.
- 👥 **Base de Clientes com Autocompletar:** Cadastre clientes frequentes para autopreenchimento instantâneo em ordens de serviço e emissão de laudos.
- 📄 **Relatórios de Serviço Profissionais:** Emissão e impressão limpa de Relatórios Técnicos, Orçamentos e Recibos de Entrega formatados para o cliente.
- 📊 **Painel Financeiro & Métricas:** Acompanhe faturamento mensal, faturamento do dia, tíquete médio, serviços em aberto e exportação de dados.
- 📱 **Progressive Web App (PWA):** Aplicativo 100% instalável em Desktop (Chrome, Edge) e dispositivos móveis (Android/iOS) com service worker para carregamento instantâneo.
- 🛡️ **Segurança & Controle de Acesso:** Autenticação baseada em JWT com validação estrita de segredo em produção, controle RBAC por perfil e log de auditoria operacional.
- 💾 **Sistema de Backup Transacional:** Snapshots atômicos e consistentes via `VACUUM INTO` do SQLite, checagem prévia de integridade (`PRAGMA integrity_check`), rotação de retenção configurável e suporte a backups pré-migração.
- 🌗 **Interface Moderna:** Totalmente responsiva com alternância entre Tema Claro e Tema Escuro.

---

## 🏗️ Arquitetura & Stack Tecnológica

O **OptiManager** foi construído com foco em **baixo consumo de recursos** (< 40MB de RAM em execução), zero dependências pesadas e facilidade máxima de implantação:

- **Backend:** [Go](https://go.dev/) (Go 1.24+), utilizando o router padrão `http.ServeMux` da standard library, middlewares nativos para CORS, gzip, logging e recovery.
- **Banco de Dados:** [SQLite](https://www.sqlite.org/) embarcado operando em modo WAL (*Write-Ahead Logging*), com versionamento por `PRAGMA user_version` e transações consistentes.
- **Frontend:** [Vue 3](https://vuejs.org/) (Composition API), [TypeScript](https://www.typescriptlang.org/), [Vite](https://vitejs.dev/), [Tailwind CSS](https://tailwindcss.com/), [Pinia](https://pinia.vuejs.org/) e [Lucide Icons](https://lucide.dev/).
- **Containerização:** [Docker](https://www.docker.com/) multi-stage build resultando em uma imagem Alpine minimalista com todos os ativos web embutidos diretamente no binário (`//go:embed`).

---

## 🚀 Como Executar com Docker (Recomendado)

### Pré-requisitos
- [Docker](https://docs.docker.com/get-docker/) e [Docker Compose](https://docs.docker.com/compose/) instalados no seu sistema.

### 1. Clonar o Repositório
```bash
git clone https://github.com/jvfigueiro/optimanager.git
cd optimanager
```

### 2. Inicializar o Arquivo de Configuração (.env)
O repositório inclui um script seguro para gerar automaticamente sua chave secreta criptográfica sem sobrescrever configurações existentes:

```bash
chmod +x scripts/init-env.sh
./scripts/init-env.sh
```

> **Nota:** O script criará o arquivo `.env` com um segredo aleatório de 64 caracteres hexadecimais (`JWT_SECRET`) gerado via `openssl rand -hex 32`.

### 3. Iniciar a Aplicação
```bash
docker compose up -d
```

### 4. Acessar o Sistema
Abra seu navegador e acesse:
```
http://localhost:8080
```
Você será recebido pelo **Assistente de Boas-Vindas**, onde definirá sua conta de Administrador e os dados da sua empresa.

---

## ⚙️ Variáveis de Ambiente

O arquivo `.env` controla o comportamento da instância:

| Variável | Descrição | Padrão | Obrigatório em Produção |
| :--- | :--- | :--- | :--- |
| `PORT` | Porta HTTP na qual o serviço escuta | `8080` | Não |
| `DB_PATH` | Caminho do arquivo de banco SQLite no container | `/data/optimanager.db` | Não |
| `TZ` | Fuso horário da aplicação | `America/Sao_Paulo` | Não |
| `JWT_SECRET` | Chave HMAC de assinatura de tokens (mínimo 32 caracteres) | *Gerada pelo script* | **Sim** |
| `APP_ENV` | Ambiente (`production` ou `dev`) | `production` | Não |
| `TRUST_PROXY` | Se `true`, respeita cabeçalhos `X-Forwarded-For` para auditoria | `false` | Não |

---

## 🛠️ Desenvolvimento Local

Caso deseje executar ou modificar o código localmente fora do Docker:

### Backend (Go)
```bash
# Executar testes unitários e de concorrência com detecção de race conditions
go test ./... -race -count=1

# Executar validação estática
go vet ./...

# Iniciar o servidor de desenvolvimento
APP_ENV=dev DB_PATH=./data/optimanager.db go run cmd/server/main.go
```

### Frontend (Vue 3 / Vite)
```bash
cd ui
npm install
npm run build   # Checagem de tipos (vue-tsc) e build de produção
npm run dev     # Servidor de desenvolvimento com HMR
```

---

## 📁 Estrutura de Diretórios

```
optimanager/
├── cmd/
│   └── server/             # Ponto de entrada do servidor Go (main.go)
├── docs/                   # Especificações técnicas e planos arquiteturais
├── internal/
│   ├── database/           # Conexão SQLite (WAL), repositórios e migrações
│   │   └── migrations/     # Scripts SQL versionados por PRAGMA user_version
│   ├── handlers/           # Handlers HTTP REST (Auth, OS, Clientes, Config, Auditoria)
│   ├── models/             # Estruturas de dados de domínio (Go structs)
│   └── services/
│       ├── backup/         # Serviço de backup atômico (VACUUM INTO, retenção, integridade)
│       └── pdf/            # Utilitários de relatórios
├── scripts/
│   └── init-env.sh         # Script utilitário para geração de chaves no .env
├── ui/                     # Aplicação frontend Single Page Application (Vue 3)
│   ├── public/             # Favicons, manifesto PWA e Service Worker
│   └── src/
│       ├── components/     # Modais, layout principal e componentes reutilizáveis
│       ├── stores/         # Gerenciamento de estado global (Pinia)
│       └── views/          # Telas (Dashboard, Serviços, Clientes, Financeiro, Relatórios, etc.)
├── Dockerfile              # Build multi-stage enxuto (Node + Go -> Alpine)
├── docker-compose.yml      # Configuração de orquestração local com volume persistente
└── .env.example            # Modelo de variáveis de ambiente
```

---

## 📄 Licença

Este projeto é distribuído sob os termos da licença **MIT**. Consulte o arquivo [LICENSE](LICENSE) para obter detalhes.
