# Módulo: Saúde e Monitoramento

## Visão geral e problema que resolve
Expõe um endpoint leve de verificação de saúde (`GET /api/health`) para confirmar que o servidor HTTP está respondendo e que o banco de dados SQLite está acessível com `journal_mode` e `foreign_keys` ativos. É utilizado pelo `HEALTHCHECK` do Docker Compose e por scripts de monitoramento.

## Como usar (passo a passo do ponto de vista do técnico)
1. Acesse via navegador ou `curl`/`wget`:
   ```bash
   curl http://localhost:8585/api/health
   ```
2. Verifique se o campo `"status"` é `"ok"` e `"database"` é `"connected"`.

## Como funciona por dentro
1. `HealthHandler.Check` (`internal/handlers/health.go`) executa duas consultas rápidas no pool de leitura (`h.db.Reader`):
   - `PRAGMA journal_mode;`
   - `PRAGMA foreign_keys;`
2. Retorna JSON contendo `status`, `app` (`"OptiManager"`), `version` (`"1.0.S7 Beta"`), `database`, `journal_mode`, `foreign_keys` e `timestamp`.
3. Se ambas as consultas tiverem sucesso, responde HTTP `200 OK`; caso contrário, responde HTTP `503 Service Unavailable`.

## Entradas, saídas e configurações
- **Endpoint:** `GET /api/health` (público).
- **Uso no Docker:** configurado no `docker-compose.yml` (`wget -qO- http://127.0.0.1:8585/api/health`, intervalo de 30s).

## Requisitos
- **Admin no sistema / Windows / Dependências externas:** nenhum.

## Riscos e efeitos colaterais
- Somente leitura; nenhum efeito colateral no banco ou sistema.

## Logs e tratamento de erros
- Requisições são registradas pelo `middleware.Logger` do `chi`. Falhas de consulta retornam `database: "error"` com HTTP `503`.

## Arquivos-fonte relacionados
- `internal/handlers/health.go`
- `cmd/server/main.go`
- `docker-compose.yml`
