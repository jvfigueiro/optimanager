# Módulo: Autenticação e Usuários

## Visão geral e problema que resolve
Controla o primeiro acesso ao sistema (Assistente de Boas-Vindas), a autenticação de operadores e administradores via token HMAC-SHA256 e o controle de acesso baseado em perfis e permissões (RBAC). Garante que apenas usuários autorizados acessem dados operacionais, financeiros ou administrativos.

## Como usar (passo a passo do ponto de vista do técnico)
1. **Primeira execução (Setup Inicial):**
   - Ao abrir o OptiManager sem usuários cadastrados, o Assistente de Boas-Vindas exibe 6 passos: Boas-Vindas, Licença (GPLv3) & Termos, Opções de Instalação (Nova ou Restaurar Backup), Conta de Administrador, Dados da Empresa (incluindo upload de logotipo) e Conclusão.
2. **Login regular:**
   - Informe e-mail e senha na tela de acesso (`AuthView.vue`). Opcionalmente marque **Lembrar meu e-mail**.
3. **Gerenciar usuários (Administrador):**
   - Acesse **Configurações → Usuários e Permissões**, clique em **Adicionar Usuário** ou edite um usuário existente para alterar nome, e-mail, senha, perfil (`admin` / `operador`) ou desativar o acesso.

## Como funciona por dentro
1. Ao iniciar o frontend (`App.vue`), `authStore.checkStatus()` consulta `GET /api/auth/status`.
2. Se `needs_setup` for `true` (0 usuários ativos no banco), a UI exibe o assistente. Ao concluir:
   - Envia `POST /api/auth/setup` com credenciais do administrador e dados da empresa (`fantasia`, `razao_social`, `cnpj`, `fones`, `email_contato`, `endereco`, `cidade`, `logo_base64`).
   - O backend cria o usuário `admin` com hash `bcrypt`, atualiza a tabela `config`, gera um token assinado com HMAC-SHA256 (validade de 7 dias) e grava o cookie `HttpOnly` `optimanager_token`.
3. Nas requisições seguintes, `authHandler.RequireAuth` extrai o token do header `Authorization: Bearer ...` ou do cookie `optimanager_token`, valida a assinatura HMAC e carrega o usuário no contexto da requisição.
4. Rotas administrativas usam `RequireAdmin`, e rotas operacionais usam `RequirePermission(perm)`.
5. Se configurado (`auto_logout_enabled = 1`), `App.vue` monitora a inatividade do usuário via `localStorage` (`optimanager_last_activity`) e executa `authStore.logout()`.

## Entradas, saídas e configurações
- **Endpoints:**
  - `GET /api/auth/status` (público)
  - `POST /api/auth/setup` (público apenas antes do primeiro admin)
  - `POST /api/auth/login` e `POST /api/auth/logout`
  - `GET /api/auth/me` (autenticado)
  - `GET/POST /api/usuarios`, `PUT/DELETE /api/usuarios/{id}` (admin)
- **Variáveis de ambiente:** `JWT_SECRET` (mínimo 32 caracteres; rejeita valores padrão inseguros) e `APP_ENV=dev` (gera chave efêmera apenas em desenvolvimento). No modo portátil sem `/data`, `ensurePortableJWTSecret` persiste a chave em `<dataDir>/.jwt_secret`.

## Requisitos
- **Admin no sistema (OptiManager):** obrigatório para listar, criar, editar ou excluir usuários.
- **Admin no Windows:** não necessário.
- **Dependências externas:** nenhuma (`golang.org/x/crypto/bcrypt` e stdlib Go).

## Riscos e efeitos colaterais
- Alterar ou perder `JWT_SECRET` invalida todas as sessões ativas, exigindo novo login.
- Desativar ou excluir um usuário impede imediatamente novos logins e invalida tokens vinculados a ele (o backend valida o status `ativo` no banco a cada requisição).
- As alterações são reversíveis via interface administrativa (exceto exclusão definitiva de usuário, que remove o registro da tabela `usuarios`).

## Logs e tratamento de erros
- Falhas na inicialização de `JWT_SECRET` interrompem o boot com `log.Fatalf("[OptiManager] Fatal authentication initialization: ...")`.
- Tokens inválidos, expirados ou de usuários inativos retornam HTTP `401 Unauthorized`; o frontend remove `optimanager_token` do `localStorage` e dispara `optimanager:unauthorized`.
- Conflitos de edição simultânea retornam HTTP `409 Conflict` (`version_conflict`).

## Arquivos-fonte relacionados
- `internal/handlers/auth.go`
- `internal/models/usuario.go`
- `internal/database/usuario_repo.go`
- `cmd/server/main.go` (`ensurePortableJWTSecret`)
- `ui/src/stores/auth.ts`
- `ui/src/views/AuthView.vue`
- `ui/src/App.vue`
