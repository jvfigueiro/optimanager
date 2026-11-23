# Configurações, Logs e Persistência de Dados

## 1. Variáveis de Ambiente

Definidas via `.env` (template em `.env.example`, inicializado com `./scripts/init-env.sh`) ou variáveis do sistema operacional:

| Variável | Padrão | Descrição |
| :--- | :--- | :--- |
| `PORT` | `8585` | Porta HTTP de escuta do servidor. |
| `DB_PATH` | `/data/optimanager.db` (Docker) ou `<exe_dir>/data/optimanager.db` (Portátil) | Caminho completo do arquivo SQLite. |
| `TZ` | `America/Sao_Paulo` | Fuso horário usado pelo container, agendador de backups e datas. |
| `JWT_SECRET` | *(obrigatório em produção)* | Chave HMAC-SHA256 com mínimo de 32 caracteres. No modo portátil (sem `/data` e sem `APP_ENV`), é gerada e salva automaticamente em `<dataDir>/.jwt_secret`. |
| `APP_ENV` | `production` | Quando `dev`, permite iniciar sem `JWT_SECRET` usando uma chave efêmera em memória. |
| `TRUST_PROXY` | `false` | Se `true`, `ExtractClientIP` lê `X-Forwarded-For` / `X-Real-IP` para logs de auditoria. |
| `BACKUP_EXTERNAL_DIR` | *(vazio)* | Diretório secundário opcional para replicação automática de backups `.zip`. |

---

## 2. Configurações no Banco e no Navegador

- **Banco de Dados (`config`):** armazena razão social, nome fantasia, CNPJ, contatos, `logo_base64`, `cor_primaria`, `cor_secundaria`, `termo_garantia`, `obs_padrao`, `custom_presets` (JSON com listas e rótulos de OS/Serviços), `auto_logout_enabled`, `auto_logout_minutos`, além de agendamento, retenção e status dos últimos backups.
- **Navegador (`localStorage`):**
  - `optimanager_token`: token de sessão JWT/HMAC.
  - `optimanager_theme`: preferência manual `light` ou `dark`.
  - `optimanager_sidebar_open`: estado recolhido/expandido da barra lateral.
  - `optimanager_remember_email` / `optimanager_saved_email`: opção de lembrar e-mail no login.
  - `optimanager_last_activity`: timestamp da última interação para auto-logout.
  - `optimanager_dismissed_recs`: IDs de recomendações dispensadas no Dashboard.

---

## 3. Logs, Auditoria e Privacidade (LGPD)

- **Logs de Servidor (stdout/stderr):**
  - Requisições HTTP registradas via `middleware.Logger` do `chi`.
  - Eventos de inicialização, banco e backup usam os prefixos `[OptiManager]`, `[DB]`, `[BACKUP WARNING]`, `[SECURITY WARNING]` e `[profile-dbf]`.
  - Em execução Docker, os logs são consultados via `docker compose logs -f optimanager`.
- **Auditoria Persistente (`audit_log`):**
  - Registra ações relevantes com `usuario_id`, `usuario_nome`, `ip`, `acao`, `detalhe`, `entidade`, `entidade_id` e `created_at`.
- **Redação Automática:**
  - `internal/redact/redact.go` (`SanitizeLogMessage`) mascara padrões de CPF, CNPJ e telefones em mensagens de erro retornadas pela API (`[REDACTED_DOC]`, `[REDACTED_PHONE]`).
  - Mais detalhes em [DADOS-E-BACKUPS.md](./DADOS-E-BACKUPS.md).
