# Módulo: Backup e Restauração

## Visão geral e problema que resolve
Protege os dados da operação contra falhas de disco, exclusões acidentais ou atualizações de schema, criando snapshots consistentes do SQLite (sem parar o servidor), compactando-os com logotipo e anexos em `.zip`, aplicando política de retenção e permitindo restauração segura.

## Como usar (passo a passo do ponto de vista do técnico)
1. **Backups em operação (Administrador):**
   - Acesse **Configurações → Backups**.
   - Clique em **Criar Backup Agora** para gerar um snapshot manual imediato.
   - Configure o agendamento automático (`diário` ou `semanal`, horário e quantidades de retenção por categoria).
   - Faça download de um `.zip`, restaure um backup listado ou envie um arquivo `.zip`/`.db` externo.
2. **Restauração no Assistente Inicial:**
   - No Passo 3 do Assistente de Boas-Vindas (`AuthView.vue`), selecione **Restaurar um backup** e envie o arquivo `.zip` ou `.db` (até 100 MB).

## Como funciona por dentro
1. **Criação do Backup (`internal/services/backup/service.go`):**
   - Verifica espaço livre em disco antes de iniciar (`disk_windows.go` via `kernel32.dll!GetDiskFreeSpaceExW` no Windows; `disk_unix.go` via `syscall.Statfs` no Linux/Unix).
   - Executa `VACUUM INTO` no SQLite para gerar um snapshot consistente em arquivo temporário `.db.tmp`.
   - Valida a integridade com `PRAGMA integrity_check`.
   - Empacota banco, logotipo (`company_logo.*`) e pasta `uploads/` em um `.zip` no padrão `optimanager_backup_YYYY-MM-DD_HH-mm-ss_<origem>.zip`.
   - Opcionalmente replica para `BACKUP_EXTERNAL_DIR` e aplica a retenção configurada (`auto`, `manual`, `prerestore`, `premigration`).
2. **Agendador:** `backupService.StartScheduler(ctx)` roda em goroutine de fundo disparando backups automáticos conforme frequência/horário salvos em `config`.
3. **Restauração:** valida o `.zip`/`.db` enviado, gera um backup `prerestore` de segurança (quando já há banco ativo), fecha os pools SQLite, remove `-wal`/`-shm`, substitui atomicamente o `.db` via `db.ReplaceDatabase` e reexecuta migrations pendentes.

## Entradas, saídas e configurações
- **Endpoints:**
  - `POST /api/auth/restore` (público apenas quando `needs_setup = true`)
  - `GET /api/backups`, `POST /api/backups`, `PUT /api/backups/config`, `POST /api/backups/restore`, `GET /api/backups/{filename}/download`, `DELETE /api/backups/{filename}` (admin)
- **Variável de ambiente:** `BACKUP_EXTERNAL_DIR` (diretório secundário opcional para cópia externa).
- **Limite de upload:** `100 MB` (`maxBackupUploadBytes` em `internal/handlers/backup.go`).

## Requisitos
- **Admin no sistema (OptiManager):** obrigatório para todas as operações em `/api/backups`.
- **Admin no Windows:** não necessário (usa permissão de escrita na pasta `data\backups` e `GetDiskFreeSpaceExW`).
- **Dependências externas:** nenhuma (stdlib `archive/zip` e SQLite `VACUUM INTO`).

## Riscos e efeitos colaterais
- Restaurar um backup **substitui integralmente** o banco de dados ativo (os dados posteriores ao snapshot são substituídos, embora um backup `prerestore` seja criado automaticamente antes da troca).
- Excluir um backup na interface remove o arquivo `.zip`/`.db` do diretório local.

## Logs e tratamento de erros
- Erros de permissão, espaço em disco insuficiente (`ErrInsufficientSpace`), nome inválido (`ErrInvalidFilename`) ou path traversal (`ErrPathTraversal`) são registrados no status do `config` (`backup_ultimo_erro`, `backup_ultimo_aviso_externo`) e logados no stdout com prefixo `[BACKUP WARNING]`.
- Arquivos `.db.tmp` e `.perm_check*` órfãos são limpos automaticamente na inicialização (`CleanupOrphanTmpFiles`).

## Arquivos-fonte relacionados
- `internal/services/backup/service.go`
- `internal/services/backup/disk_windows.go`
- `internal/services/backup/disk_unix.go`
- `internal/handlers/backup.go`
- `internal/database/db.go` (`ReplaceDatabase`)
- `ui/src/views/ConfiguracoesView.vue`
- `ui/src/views/AuthView.vue`
