# Módulo: Auditoria e Privacidade (LGPD)

## Visão geral e problema que resolve
Registra ações administrativas e operacionais sensíveis para rastreabilidade (`audit_log`), protege dados pessoais em mensagens de erro e logs do servidor (redação automática de CPF, CNPJ e telefones) e garante identificação correta do IP do cliente atrás de proxies reversos.

## Como usar (passo a passo do ponto de vista do técnico)
- Este módulo atua de forma transparente em segundo plano durante o uso normal do sistema.
- Para ambientes atrás de proxy reverso (Nginx, Caddy, Traefik), configure `TRUST_PROXY=true` no `.env` para que o IP real do cliente seja registrado no `audit_log`.

## Como funciona por dentro
1. **Registro de Auditoria:** handlers de criação/edição/exclusão/restauração gravam eventos na tabela `audit_log` via `internal/database/audit_repo.go` (`usuario_id`, `usuario_nome`, `ip`, `acao`, `detalhe`, `entidade`, `entidade_id`, `created_at`).
2. **Redação LGPD (`internal/redact/redact.go`):**
   - `SanitizeLogMessage` aplica expressões regulares para substituir CPFs/CNPJs formatados (`[REDACTED_DOC]`), sequências numéricas de 11 ou 14 dígitos (`[REDACTED_DOC]`) e telefones brasileiros (`[REDACTED_PHONE]`).
   - Todas as respostas de erro HTTP enviadas via `respondError` (`internal/handlers/response.go`) passam por `SanitizeLogMessage`.
3. **Extração de IP (`internal/handlers/request_ip.go`):**
   - Se `TRUST_PROXY=true`, `ExtractClientIP` lê o primeiro salto de `X-Forwarded-For` ou `X-Real-IP`. Caso contrário, utiliza exclusivamente `r.RemoteAddr` para evitar spoofing de cabeçalhos.

## Entradas, saídas e configurações
- **Tabela:** `audit_log` (com índice `idx_audit_log_created_at`).
- **Variável de ambiente:** `TRUST_PROXY` (`true` ou `false`, padrão `false`).

## Requisitos
- **Admin no sistema / Windows / Dependências externas:** nenhum (integrado ao núcleo do servidor).

## Riscos e efeitos colaterais
- Se `TRUST_PROXY=true` for habilitado sem um proxy reverso confiável na frente da aplicação, clientes poderão forjar o cabeçalho `X-Forwarded-For` nos registros de auditoria.

## Logs e tratamento de erros
- Mensagens de erro retornadas pela API têm documentos e telefones automaticamente mascarados antes de serem serializadas em JSON.

## Arquivos-fonte relacionados
- `internal/database/audit_repo.go`
- `internal/models/audit_log.go`
- `internal/redact/redact.go`
- `internal/handlers/request_ip.go`
- `internal/handlers/response.go`
