# Módulo: Financeiro

## Visão geral e problema que resolve
Consolida receitas, valores pendentes de recebimento, distribuição por forma de pagamento e métricas diárias/mensais a partir das Ordens de Serviço, permitindo impressão de resumo e exportação em CSV.

## Como usar (passo a passo do ponto de vista do técnico)
1. Acesse **Financeiro** na barra lateral.
2. Utilize os filtros de período/mês e busca para analisar as ordens faturadas e pendentes.
3. Clique em **Exportar CSV** para baixar uma planilha com os lançamentos filtrados ou use a impressão (`window.print()`, área `#financeiro-print`).

## Como funciona por dentro
1. `FinanceiroView.vue` reutiliza `ordensStore.ordens` e `configStore.config` (sem endpoint HTTP separado exclusivo para o painel financeiro).
2. Os cálculos de faturamento recebido, a receber, ticket médio e agrupamentos por forma de pagamento são computados reativamente no frontend a partir da lista de ordens carregada pelo backend.
3. Se o usuário autenticado não tiver a permissão `ver_valores` no backend, `OrdemHandler` aplica `RedactValues()` antes de enviar os dados para o navegador.
4. A impressão utiliza regras `@media print` definidas em `ui/src/style.css` isolando `#financeiro-print`.

## Entradas, saídas e configurações
- **Fonte de dados:** `GET /api/ordens` e `GET /api/config`.
- **Saídas:** visualização em tela, impressão A4 (`#financeiro-print`) e download de arquivo `.csv` gerado no navegador.

## Requisitos
- **Admin no sistema (OptiManager):** não obrigatório, desde que o usuário possua acesso às ordens e permissão `ver_valores` / `ver_financeiro` conforme perfil.
- **Admin no Windows / Dependências externas:** nenhum.

## Riscos e efeitos colaterais
- Módulo somente leitura sobre os dados já carregados de Ordens de Serviço; não altera o banco de dados.
- Exportações CSV contêm dados operacionais e devem ser armazenadas com cuidado pelo operador.

## Logs e tratamento de erros
- Falhas ao carregar ordens são tratadas em `ordensStore` e exibidas via toast.

## Arquivos-fonte relacionados
- `ui/src/views/FinanceiroView.vue`
- `ui/src/stores/ordens.ts`
- `ui/src/utils/formatters.ts`
- `ui/src/style.css`
