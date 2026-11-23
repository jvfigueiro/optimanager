# Mapa de Arquitetura e Inventário — OptiManager (`1.0.S7 Beta`)

> Análise estrita baseada no código-fonte atual do repositório. Itens que dependem de ambiente externo ou não puderam ser verificados diretamente no código estão marcados com `[VERIFICAR]`.

---

## 1. Stack Tecnológica

### Backend & Runtime
- **Linguagem:** Go `1.26.1` (`go.mod`).
- **Runtime:** Binário nativo estático (`CGO_ENABLED=0`), multiplataforma (Linux amd64 e Windows amd64 portátil), sem dependências externas em tempo de execução.
- **Roteamento HTTP e Middlewares:** `github.com/go-chi/chi/v5 v5.5.0` — roteador REST, `RequestID`, `RealIP`, `Logger`, `Recoverer`, `Timeout(120s)` e CORS customizado (`cmd/server/main.go`).
- **Banco de Dados:** `modernc.org/sqlite v1.61.0` — driver SQLite em Go puro (sem CGO), configurado em modo WAL (*Write-Ahead Logging*), `busy_timeout(5000)`, `synchronous(NORMAL)`, `foreign_keys(ON)` e `cache_size(-8000)` (`internal/database/db.go`).
- **Geração de PDF:** `github.com/johnfercher/maroto/v2 v2.4.4` — geração de documentos A4 em PDF no backend (`GET /api/ordens/{id}/pdf`, `internal/services/pdf/pdf_service.go`).
- **Criptografia:** `golang.org/x/crypto v0.58.0` (`bcrypt` para hash de senhas em `internal/handlers/auth.go` e `internal/database/usuario_repo.go`), além da stdlib `crypto/hmac`, `crypto/sha256` e `crypto/rand` para tokens de sessão e geração de `JWT_SECRET`.
- **Conversão de Texto Legado:** `golang.org/x/text v0.43.0` — decodificação de encodings legados (`cp1252`, `cp850`, `cp437`) no leitor DBF (`internal/legacy/dbf`).
- **Timezone embutido:** `_ "time/tzdata"` (stdlib) — garante resolução de `America/Sao_Paulo` mesmo no executável portátil Windows ou imagens Alpine mínimas.

### Frontend
- **Framework:** Vue `3.5.13` (Composition API, `<script setup lang="ts">`).
- **Linguagem:** TypeScript `5.7.3` (`vue-tsc --noEmit`).
- **Gerenciamento de Estado:** Pinia `2.3.1` (`ui/src/stores/`).
- **Estilização:** Tailwind CSS `3.4.17` + PostCSS `8.5.2` + Autoprefixer `10.4.20`.
- **Ícones:** `lucide-vue-next ^0.475.0`.
- **Empacotamento Web:** Vite `6.1.0` (`@vitejs/plugin-vue ^5.2.1`).
- **PWA:** `ui/public/manifest.json`, `manifest.webmanifest` e Service Worker (`ui/public/sw.js`).
- **Embedding no Binário:** `//go:embed all:dist` em `ui/embed.go` serve a SPA compilada diretamente pelo servidor Go.

### Ferramentas de Build, Empacotamento e Setup
- **Makefile:** alvos `build-ui`, `build-windows`, `build-linux`, `test` e `clean`.
- **Docker & Docker Compose:** `Dockerfile` multi-stage (`node:22-alpine` → `golang:alpine` → `alpine:3.21`) e `docker-compose.yml` com volume `/data` e healthcheck em `/api/health`.
- **Scripts Shell:** `scripts/build-windows.sh` (build UI + cross-compile Windows), `scripts/init-env.sh` (gera `.env` com `JWT_SECRET`), `scripts/run-demo.sh`, `scripts/seed-demo.sh`, `scripts/check-forbidden-files.sh`, `scripts/install-hooks.sh`.
- **Script Windows (PowerShell):** `scripts/windows/setup-optimanager.ps1` (copiado para `dist/` no `make build-windows`), que cria regra de Firewall e Tarefa Agendada no Windows. Não há gerador de instalador MSI/NSIS/InnoSetup no repositório; a distribuição Windows é portátil (`optimanager-windows-amd64.exe` + script `.ps1`).

---

## 2. Estrutura de Pastas e Responsabilidades

```text
optimanager/
├── cmd/server/                  # Ponto de entrada: CLI `profile-dbf` e servidor HTTP/SPA
├── internal/
│   ├── database/                # Pool SQLite (1 writer / 8 readers), migrations SQL e repositórios
│   │   └── migrations/          # Scripts SQL numerados (001 a 004) controlados por PRAGMA user_version
│   ├── handlers/                # Controladores HTTP REST, autenticação JWT/Cookie, RBAC e IP extraction
│   ├── legacy/dbf/              # Leitor e profiler de arquivos DBF legados (CLI `profile-dbf`)
│   ├── models/                  # Structs de domínio, DTOs e constantes de permissões RBAC
│   ├── money/                   # Conversão segura entre reais (float64) e centavos (int64)
│   ├── redact/                  # Sanitização LGPD de CPF, CNPJ e telefones em mensagens de erro/log
│   └── services/
│       ├── backup/              # Motor de backup atômico (VACUUM INTO, ZIP, retenção, agendador, espaço em disco)
│       └── pdf/                 # Gerador de PDF A4 (Maroto v2) para OS/Relatório/Orçamento/Recibo
├── ui/                          # Frontend SPA (Vue 3 + TypeScript + Vite + Tailwind + Pinia)
│   ├── embed.go                 # Diretiva //go:embed all:dist
│   ├── public/                  # Ícones, manifests PWA e Service Worker
│   └── src/
│       ├── api/                 # Cliente HTTP Fetch centralizado (`api/index.ts`)
│       ├── components/          # Layout principal (`AppLayout.vue`), modais de OS/Atalhos e Toasts
│       ├── stores/              # Stores Pinia (auth, clientes, config, dashboard, navigation, ordens, toast)
│       ├── types/               # Interfaces TypeScript espelhando DTOs da API
│       ├── utils/               # Formatadores (BRL, data, CPF/CNPJ, telefone fixo/celular, obs/problema)
│       └── views/               # Telas principais (Auth, Dashboard, Serviços, Clientes, Financeiro, Relatórios, Configurações)
├── scripts/                     # Scripts de build, ambiente, demo, hooks e setup Windows (`scripts/windows/`)
├── docs/                        # Documentação técnica e operacional
├── data/                        # Dados locais de desenvolvimento/demo (ex.: `data/demo/`)
├── dist/                        # Artefatos gerados de distribuição Windows
├── Dockerfile / docker-compose.yml / Makefile
```

---

## 3. Como a Arquitetura Modular Funciona

O OptiManager **não** utiliza um sistema de carregamento dinâmico de plugins (DLLs, `.so` ou módulos descobertos em tempo de execução). A modularidade é **arquitetural e em tempo de compilação**, estruturada em camadas verticais coesas:

1. **Definição de um Módulo:**
   - **Domínio:** structs e constantes em `internal/models/<modulo>.go`.
   - **Persistência:** métodos acoplados a `*database.DB` em `internal/database/<modulo>_repo.go`, utilizando `d.Writer` (pool de 1 conexão para escrita) e `d.Reader` (pool de até 8 conexões para leitura) e migrações SQL embutidas em `internal/database/migrations/`.
   - **Serviço (quando há lógica de negócio complexa/assíncrona):** pacote dedicado em `internal/services/<servico>/`.
   - **Transporte HTTP:** struct `<Modulo>Handler` em `internal/handlers/<modulo>.go` recebendo `*database.DB` (e serviços opcionais).
   - **Frontend:** tipos em `ui/src/types/index.ts`, métodos em `ui/src/api/index.ts`, store reativa em `ui/src/stores/<modulo>.ts` e interface em `ui/src/views/<Modulo>View.vue` ou `ui/src/components/<Modal>.vue`.

2. **Descoberta, Carregamento e Registro:**
   - **Backend:** O registro é explícito e estático em `cmd/server/main.go`:
     1. `database.New("")` abre os pools SQLite e executa as migrations pendentes.
     2. Serviços (como `backup.NewService`) e handlers (`handlers.New<Modulo>Handler(db)`) são instanciados manualmente.
     3. As rotas são montadas no roteador `chi` sob `/api` com middlewares de autenticação (`RequireAuth`), perfil (`RequireAdmin`) ou permissão granular (`RequirePermission`).
     4. A SPA compilada em `ui/dist` é embutida via `//go:embed all:dist` e servida na rota fallback `/*`.
   - **Frontend:** `ui/src/App.vue` inicializa tema, rota por hash (`nav.initFromHash()`), configurações (`configStore.loadConfig()`) e sessão (`authStore.checkStatus()`). As views são alternadas reativamente em `AppLayout.vue` com base em `nav.currentView` (`dashboard | orders | clients | finance | reports | settings`) — sem `vue-router`, usando hash da URL (`#dashboard`, `#orders`, etc.).

3. **Contrato / Interface:**
   - Não há uma interface Go única obrigatória (`Module` interface) para handlers; cada handler segue a convenção `New<Nome>Handler(db *database.DB, ...) *<Nome>Handler` com métodos padrão `http.HandlerFunc` (`w http.ResponseWriter, r *http.Request`), usando os helpers compartilhados `respondJSON`, `respondError` e `respondVersionConflict` (`internal/handlers/response.go`).
   - No frontend, os módulos consomem o cliente `request<T>()` de `ui/src/api/index.ts`, que injeta `Authorization: Bearer <token>`, `credentials: 'include'`, trata `401` (evento `optimanager:unauthorized`) e `409` (`optimanager:version-conflict`).

4. **Ciclo de Vida:**
   - Inicialização síncrona do banco e migrações → início do agendador de backups (`backupService.StartScheduler(ctx)`) → servidor HTTP ativo na porta configurada (`PORT`, padrão `8585`) → encerramento gracioso em `SIGINT`/`SIGTERM` com timeout de 5 segundos e `db.Close()`.

---

## 4. Inventário Completo de Módulos e Recursos

| Módulo | Objetivo | Arquivos Principais | Dependências Internas | Exige Admin? |
| :--- | :--- | :--- | :--- | :--- |
| **Autenticação & Usuários** | Setup inicial (6 passos), login/logout HMAC-SHA256, RBAC por perfil/permissão e CRUD de usuários | `internal/handlers/auth.go`, `internal/models/usuario.go`, `internal/database/usuario_repo.go`, `ui/src/stores/auth.ts`, `ui/src/views/AuthView.vue` | `database`, `models` | Setup/Login: Não. Gestão de usuários (`/api/usuarios`): **Sim (`admin`)** |
| **Dashboard & Atalhos** | Métricas resumidas, recomendações operacionais, ações rápidas e atalhos externos | `internal/handlers/dashboard.go`, `internal/handlers/atalho.go`, `internal/database/dashboard_repo.go`, `internal/database/atalho_repo.go`, `ui/src/views/DashboardView.vue`, `ui/src/stores/dashboard.ts` | `database`, `models` | Leitura: Não. CRUD de atalhos (`/api/atalhos`): **Sim (`admin`)** |
| **Ordens de Serviço (Serviços)** | Criação, edição, conclusão, duplicação, exclusão, ações em lote e baixa de pagamento de OS | `internal/handlers/ordem_servico.go`, `internal/database/os_repo.go`, `internal/models/ordem_servico.go`, `internal/money/money.go`, `ui/src/views/ServicosView.vue`, `ui/src/components/NovaOSModal.vue`, `ui/src/components/EditarOSModal.vue`, `ui/src/stores/ordens.ts` | `database`, `models`, `money`, `services/pdf` | Leitura/Criação/Edição: Não. Baixa: `ver_financeiro`. Exclusão: `excluir` |
| **Clientes** | Cadastro de clientes com busca, autocompletar em OS, ações em lote e link para WhatsApp | `internal/handlers/cliente.go`, `internal/database/cliente_repo.go`, `internal/models/cliente.go`, `ui/src/views/ClientesView.vue`, `ui/src/stores/clientes.ts` | `database`, `models` | Leitura: Não. Escrita: `gerenciar_clientes`. Exclusão/Lote: `excluir` |
| **Financeiro** | Análise de faturamento mensal/diário, fluxo de recebimentos, resumo para impressão e exportação CSV | `ui/src/views/FinanceiroView.vue`, `ui/src/stores/ordens.ts`, `ui/src/utils/formatters.ts` | `ordens`, `config`, `auth` | Depende de dados financeiros (redigidos no backend sem `ver_valores`) |
| **Relatórios & Geração de PDF** | Gerador de Documentos na UI (OS, Orçamento, Relatório, Recibo monocromáticos A4 + WhatsApp) e PDF nativo via Maroto v2 | `ui/src/views/RelatoriosView.vue`, `internal/services/pdf/pdf_service.go`, `internal/handlers/ordem_servico.go` (`DownloadPDF`) | `database`, `models`, `ordens`, `config` | Não |
| **Configurações** | Dados da empresa, logo, cores de destaque (accent colors), presets de OS/listas, auto-logout e tela Sobre | `internal/handlers/config.go`, `internal/database/config_repo.go`, `internal/models/config.go`, `ui/src/views/ConfiguracoesView.vue`, `ui/src/stores/config.ts` | `database`, `models` | Leitura (`GET /api/config`): Pública. Alteração (`PUT /api/config`): **Sim (`admin`)** |
| **Backup & Restauração** | Backup atômico (`VACUUM INTO`), pacote `.zip` (DB + logo + uploads), agendador automático, retenção e restauração | `internal/services/backup/service.go`, `disk_windows.go`, `disk_unix.go`, `internal/handlers/backup.go`, `ui/src/views/ConfiguracoesView.vue` | `database`, `models` | Restauração inicial (`/api/auth/restore`): Pública (apenas se `needs_setup`). Demais rotas `/api/backups`: **Sim (`admin`)** |
| **Auditoria & Privacidade (LGPD)** | Registro de ações sensíveis em `audit_log`, mascaramento/redação de CPF/CNPJ/telefone em logs e erros | `internal/database/audit_repo.go`, `internal/models/audit_log.go`, `internal/redact/redact.go`, `internal/handlers/request_ip.go`, `internal/handlers/response.go` | `database`, `models` | Interno / transparente |
| **Migração Legada DBF** | Subcomando CLI `profile-dbf` para inspecionar esquemas e encodings de arquivos `.dbf` legados | `cmd/server/main.go` (`runProfileDBF`), `internal/legacy/dbf/dbf.go`, `internal/legacy/dbf/profiler.go` | `golang.org/x/text` | CLI local (sem HTTP) |
| **Saúde & Monitoramento** | Endpoint de status do servidor e integridade dos pragmas SQLite | `internal/handlers/health.go` | `database` | Não (pública em `GET /api/health`) |

---

## 5. Pontos de Integração com o Windows

Diferentemente de ferramentas utilitárias de sistema, o OptiManager é uma aplicação web em Go + Vue.js. Suas integrações reais com o Windows são pontuais:

1. **API Win32 (`kernel32.dll`):**
   - Arquivo `internal/services/backup/disk_windows.go` (`//go:build windows`) carrega `kernel32.dll` e chama `GetDiskFreeSpaceExW` via `syscall.NewLazyDLL` para verificar o espaço livre em disco antes de executar backups.
   - Em sistemas Unix/Linux, `disk_unix.go` (`//go:build !windows`) usa `syscall.Statfs`.
2. **Executável Portátil e Caminho de Dados:**
   - Compilado com `CGO_ENABLED=0 GOOS=windows GOARCH=amd64` (`Makefile`, `scripts/build-windows.sh`).
   - `database.ResolveDataDir()` detecta ambiente portátil (quando `DB_PATH` não está definido e `/data` não existe) e resolve `<diretório_do_exe>\data`, gravando ali `optimanager.db`, `backups\`, `uploads\` e `.jwt_secret`.
3. **PowerShell (`scripts/windows/setup-optimanager.ps1`):**
   - Exige **PowerShell como Administrador** (`Test-IsAdmin`).
   - **Firewall do Windows:** `Get-NetFirewallRule` / `New-NetFirewallRule` libera entrada TCP na porta configurada (padrão `8585`) para os perfis `Private,Domain`.
   - **Agendador de Tarefas:** `Register-ScheduledTask` / `Start-ScheduledTask` registra a tarefa `OptiManager` executando como conta `SYSTEM` (`RunLevel Highest`) na inicialização (`-AtStartup`) com reinício automático em falhas.
   - **Desinstalação (`-Uninstall`):** remove a regra de Firewall e a Tarefa Agendada mantendo a pasta `data\`.
4. **NSSM (Documentado, não embutido):**
   - `docs/WINDOWS-PORTABLE.md` documenta o uso opcional do NSSM para rodar o `.exe` como serviço do Windows.
5. **Ausência de integrações nativas extras:**
   - O código **não** manipula Registro do Windows, WMI/CIM, `winget`, CMD nem possui ramificações condicionais entre Windows 10 e Windows 11. Os cmdlets `NetSecurity` (`New-NetFirewallRule`) e `ScheduledTasks` (`Register-ScheduledTask`) funcionam de forma equivalente no Windows 10 e Windows 11 `[VERIFICAR em ambiente Windows 11 real]`.

---

## 6. Configuração, Logs, Persistência de Dados e Tratamento de Erros

- **Configuração:**
  - **Variáveis de ambiente:** `PORT` (`8585`), `DB_PATH` (`/data/optimanager.db` no Docker ou `<exe_dir>/data/optimanager.db` no portátil), `TZ` (`America/Sao_Paulo`), `JWT_SECRET` (mínimo 32 chars; auto-gerado em `.jwt_secret` no modo portátil ou efêmero se `APP_ENV=dev`), `APP_ENV`, `TRUST_PROXY`, `BACKUP_EXTERNAL_DIR`.
  - **Banco de dados (`config`):** dados da empresa, logotipo (`logo_base64`), cores (`cor_primaria`, `cor_secundaria`), presets JSON (`custom_presets`), auto-logout, agendamento e retenção de backups.
  - **Navegador (`localStorage`):** `optimanager_token`, `optimanager_theme`, `optimanager_sidebar_open`, `optimanager_remember_email`, `optimanager_saved_email`, `optimanager_last_activity`, `optimanager_dismissed_recs`.
- **Logs:**
  - **Stdout/Stderr:** via pacote padrão `log` do Go e `middleware.Logger` do `chi` (requisições HTTP). Prefixos usados: `[OptiManager]`, `[DB]`, `[BACKUP WARNING]`, `[SECURITY WARNING]`, `[profile-dbf]`. Não há rotação de arquivos de log embutida na aplicação.
  - **Auditoria persistente:** tabela `audit_log` no SQLite registra ações críticas (`usuario_id`, `usuario_nome`, `ip`, `acao`, `detalhe`, `entidade`, `entidade_id`, `created_at`).
- **Persistência de Dados:**
  - SQLite em modo WAL com pool separado (`1` escritor, `8` leitores), valores monetários armazenados em centavos (`int64`), controle de concorrência otimista via coluna `versao` em `config`, `clientes`, `ordens_servico` e `usuarios`.
  - Migrações versionadas (`001` a `004`) via `PRAGMA user_version`, com backup automático `premigration` antes de evoluir o schema.
- **Tratamento de Erros:**
  - **Backend:** `respondError` sanitiza mensagens com `redact.SanitizeLogMessage` (removendo CPF, CNPJ e telefones) antes de responder JSON `{ "error": "..." }`. Conflitos de versão retornam HTTP `409 Conflict` com `code: "version_conflict"`. `middleware.Recoverer` evita queda do processo em panics.
  - **Frontend:** `ui/src/api/index.ts` intercepta `401` (desloga e dispara `optimanager:unauthorized`) e `409` (dispara `optimanager:version-conflict` abrindo modal de recarregamento em `App.vue`); stores exibem mensagens via `ToastContainer.vue`.

---

## 7. Lista de Itens Não Determinados (`[VERIFICAR]`)

1. `[VERIFICAR]` Comportamento prático do script `scripts/windows/setup-optimanager.ps1` e do binário `optimanager-windows-amd64.exe` em execuções reais no Windows 10 e Windows 11 (este ambiente de análise é Linux).
2. `[VERIFICAR]` Se há intenção futura de implementar módulos de `estoque`, `compras` e `vendas` no backend/frontend: a migration `003_foundation.sql` cria a coluna `modulos_ativos` com `{"estoque":true,"compras":true,"vendas":false,"financeiro":true}`, mas não existem tabelas, handlers ou views para `estoque`, `compras` ou `vendas` no código atual.
3. `[VERIFICAR]` Se o endpoint `GET /api/ordens/{id}/pdf` (Maroto v2) continua sendo usado por clientes externos ou integrações, já que a interface web (`RelatoriosView.vue`) gera a impressão via HTML/CSS (`window.print()`).
