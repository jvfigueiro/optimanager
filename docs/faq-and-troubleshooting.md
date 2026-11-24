# FAQ e Solução de Problemas (Troubleshooting)

Baseado nos tratamentos de erro e validações implementados diretamente no código-fonte.

---

### 1. O servidor não inicia e exibe erro sobre `JWT_SECRET`
**Causa:** `ResolveJWTSecret()` (`internal/handlers/auth.go`) exige que `JWT_SECRET` tenha pelo menos 32 caracteres e rejeita explicitamente a chave antiga padrão (`optimanager-default-secret-key-2026-hmac-sha256`) e o placeholder do `.env.example` (`CONFIGURE_UMA_CHAVE...`).
**Solução:**
- Em Docker/Linux: execute `./scripts/init-env.sh` para gerar um segredo seguro de 64 caracteres no `.env`.
- Em desenvolvimento local: defina `APP_ENV=dev` para gerar chave efêmera.
- No executável portátil (sem `/data` e sem `APP_ENV`): o servidor gera automaticamente `.jwt_secret` na pasta `data/`.

---

### 2. Erro `"Este registro foi alterado por outra pessoa"` (`409 Conflict` / `version_conflict`)
**Causa:** O OptiManager usa **concorrência otimista** (`versao`) em `config`, `clientes`, `ordens_servico` e `usuarios`. Dois operadores ou abas editaram o mesmo registro simultaneamente.
**Solução:** Clique em **Recarregar** no modal exibido por `App.vue` para obter a versão mais recente antes de salvar novamente.

---

### 3. Erro ao excluir um cliente
**Causa:** A tabela `ordens_servico` possui `FOREIGN KEY (cliente_id) REFERENCES clientes(id) ON DELETE RESTRICT`. Clientes que possuem ao menos uma Ordem de Serviço não podem ser apagados.
**Solução:** Mantenha o cadastro ou remova antes as Ordens de Serviço associadas, caso realmente deseje excluir o cliente.

---

### 4. Banco de dados corrompido ou travado em rede Windows (SMB)
**Causa:** O SQLite está configurado em modo `WAL` (`_pragma=journal_mode(WAL)`), que exige bloqueios locais do sistema de arquivos e memória compartilhada.
**Solução:** Nunca abra o arquivo `optimanager.db` a partir de uma pasta compartilhada SMB por múltiplos executáveis. Rode apenas **uma instância** no servidor e acesse dos demais dispositivos pelo navegador em `http://<IP-DO-SERVIDOR>:8585`.

---

### 5. Falha no upload ou restauração de backup
**Causas possíveis (validadas em `internal/handlers/backup.go` e `internal/services/backup/service.go`):**
- Arquivo maior que **100 MB** (`maxBackupUploadBytes`).
- Nome de arquivo fora do padrão ou tentativa de *path traversal* (`ErrInvalidFilename`, `ErrPathTraversal`).
- Espaço insuficiente em disco (`ErrInsufficientSpace`, verificado antes do `VACUUM INTO`).
- Banco dentro do `.zip`/`.db` falhou no `PRAGMA integrity_check`.
**Solução:** Verifique o espaço em disco na unidade de `data/` e a mensagem detalhada em **Configurações → Backups** (`backup_ultimo_erro`).

---

### 6. `setup-optimanager.ps1` falha ao executar no Windows
**Causas comuns:**
- PowerShell não foi aberto **como Administrador** (`Test-IsAdmin` retorna erro).
- Política de execução bloqueando scripts locais.
**Solução:** Abra o PowerShell como Administrador na pasta do executável e rode:
```powershell
powershell -ExecutionPolicy Bypass -File .\setup-optimanager.ps1
```

---

### 7. Sessão encerra sozinha após alguns minutos
**Causa:** O recurso de **Auto-Logout por Inatividade** (`auto_logout_enabled = 1`) está ativo em Configurações, monitorado em `App.vue` via `optimanager_last_activity`.
**Solução:** Ajuste o tempo em minutos ou desative o auto-logout em **Configurações**.
