# Módulo: Clientes

## Visão geral e problema que resolve
Mantém um diretório centralizado de clientes (nome/razão social, CPF/CNPJ e telefone/WhatsApp) para evitar redigitação em novas Ordens de Serviço e agilizar o contato direto via WhatsApp.

## Como usar (passo a passo do ponto de vista do técnico)
1. Acesse **Clientes** no menu lateral.
2. Clique em **Novo Cliente** para cadastrar nome/razão social, CPF/CNPJ e contato.
3. Utilize a barra de busca para filtrar por nome ou documento.
4. Na listagem, clique no ícone de **WhatsApp** para iniciar uma conversa com o cliente, ou no ícone de edição para atualizar dados.
5. Para exclusão múltipla, marque os checkboxes desejados e utilize a ação em lote.

## Como funciona por dentro
1. `ClientesView.vue` consome `clientesStore` (`ui/src/stores/clientes.ts`), que comunica com `/api/clientes`.
2. `ClienteHandler` delega consultas e mutações para `internal/database/cliente_repo.go`.
3. A tabela `clientes` possui chave estrangeira referenciada por `ordens_servico.cliente_id` com `ON DELETE RESTRICT`, impedindo exclusão acidental de clientes que possuem OS vinculadas.
4. Atualizações utilizam a coluna `versao` para controle de concorrência otimista.

## Entradas, saídas e configurações
- **Endpoints:**
  - `GET /api/clientes`, `GET /api/clientes/{id}`
  - `POST /api/clientes`, `PUT /api/clientes/{id}` (permissão `gerenciar_clientes`)
  - `DELETE /api/clientes/{id}`, `POST /api/clientes/batch` (permissão `excluir`)
- **Link externo:** botão de WhatsApp abre `https://wa.me/55<numero>` no navegador.

## Requisitos
- **Admin no sistema (OptiManager):** criar/editar requer permissão `gerenciar_clientes`; excluir requer `excluir`.
- **Admin no Windows / Dependências externas:** nenhum.

## Riscos e efeitos colaterais
- Clientes vinculados a Ordens de Serviço existentes não podem ser excluídos diretamente devido à restrição `FOREIGN KEY ... ON DELETE RESTRICT`.
- Exclusões de clientes sem OS vinculadas são permanentes no banco ativo.

## Logs e tratamento de erros
- Erros de integridade referencial, validação ou conflito de versão (`409`) são sanitizados por `respondError` e exibidos em toast no frontend.

## Arquivos-fonte relacionados
- `internal/handlers/cliente.go`
- `internal/database/cliente_repo.go`
- `internal/models/cliente.go`
- `ui/src/views/ClientesView.vue`
- `ui/src/stores/clientes.ts`
