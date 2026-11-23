# Arquitetura do OptiManager

O **OptiManager** (`1.0.S7 Beta`) é uma aplicação web auto-hospedada construída como um **monólito modular em binário único**: o backend em Go (`cmd/server`) compila e embute o frontend SPA Vue 3 (`ui/dist` via `//go:embed all:dist`) e utiliza SQLite embarcado em modo WAL.

## Visão Geral das Camadas

1. **Camada de Apresentação (`ui/src/`):** SPA em Vue 3 + TypeScript + Tailwind CSS + Pinia. A navegação entre telas ocorre reativamente via store `navigation.ts` e hash da URL (`#dashboard`, `#orders`, `#clients`, `#finance`, `#reports`, `#settings`), sem `vue-router`.
2. **Camada de Transporte HTTP (`internal/handlers/`):** Roteador `chi/v5` expondo a API REST sob `/api/*`, com middlewares de `RequestID`, `RealIP`, `Logger`, `Recoverer`, `Timeout(120s)`, CORS e autenticação JWT/Cookie HMAC-SHA256 (`RequireAuth`, `RequireAdmin`, `RequirePermission`).
3. **Camada de Serviços (`internal/services/`):** Lógica de domínio especializada e processos em segundo plano — motor de backup (`backup`) com agendador e geração de PDF (`pdf` via Maroto v2).
4. **Camada de Persistência (`internal/database/`):** Repositórios SQLite usando dois pools separados (`Writer`: 1 conexão; `Reader`: 8 conexões), migrações SQL sequenciais controladas por `PRAGMA user_version` e concorrência otimista (`versao`).
5. **Utilitários Transversais (`internal/models/`, `internal/money/`, `internal/redact/`, `internal/legacy/dbf/`):** Entidades e permissões RBAC, aritmética em centavos, redação LGPD em erros/logs e profiler CLI de arquivos DBF.

## Diagrama da Arquitetura Modular

```mermaid
flowchart TB
    Browser["Navegador / PWA (Vue 3 + Pinia + Tailwind)"]
    Embed["//go:embed all:dist (ui/embed.go)"]
    Router["Servidor HTTP Go + chi/v5 (cmd/server/main.go)"]

    subgraph Middlewares ["Middlewares & Segurança"]
        MW["RequestID / RealIP / Logger / Recoverer / CORS"]
        AuthMW["RequireAuth / RequireAdmin / RequirePermission"]
        Redact["redact.SanitizeLogMessage (LGPD)"]
    end

    subgraph Handlers ["Módulos HTTP (internal/handlers)"]
        HAuth["AuthHandler"]
        HDash["DashboardHandler & AtalhoHandler"]
        HOS["OrdemHandler"]
        HCli["ClienteHandler"]
        HCfg["ConfigHandler"]
        HBkp["BackupHandler"]
        HHlt["HealthHandler"]
    end

    subgraph Services ["Serviços (internal/services)"]
        SBkp["backup.Service + Scheduler + disk_windows/unix"]
        SPDF["pdf.GenerateDocument (Maroto v2)"]
    end

    subgraph Persistence ["Persistência (internal/database)"]
        DBPools["SQLite WAL (Writer: 1 | Reader: 8)"]
        Repos["Repos: usuario, os, cliente, config, dashboard, atalho, audit"]
        Migrations["Migrations SQL (001..004)"]
    end

    Browser -->|HTTP /api/*| Router
    Browser -->|GET /* SPA| Embed
    Router --> MW --> AuthMW --> Handlers
    Handlers --> Redact
    HOS --> SPDF
    HBkp --> SBkp
    Handlers --> Repos
    SBkp --> DBPools
    Repos --> DBPools
    Migrations --> DBPools
```

## Fluxo de Inicialização e Carregamento

```mermaid
sequenceDiagram
    participant Main as cmd/server/main.go
    participant DB as internal/database
    participant Bkp as services/backup
    participant HTTP as chi Router (:8585)
    participant UI as Vue App (App.vue)

    Main->>DB: database.New("") (cria pasta data/, abre pools Writer/Reader)
    DB->>DB: migrate() (PRAGMA user_version + backup premigration se necessário)
    Main->>Main: ensurePortableJWTSecret() (gera .jwt_secret se modo portátil)
    Main->>Bkp: backup.NewService() + CleanupOrphanTmpFiles()
    Main->>Bkp: StartScheduler(ctx) (goroutine de backups automáticos)
    Main->>HTTP: Registra rotas /api/* e fallback /* (ui.DistFS)
    UI->>HTTP: GET /api/config + GET /api/auth/status
    alt needs_setup == true
        UI->>UI: Exibe Assistente de Boas-Vindas (AuthView.vue)
    else authenticated == true
        UI->>HTTP: Carrega clientes, dashboard e exibe AppLayout.vue
    end
```
