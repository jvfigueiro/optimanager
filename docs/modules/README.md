# Índice de Módulos — OptiManager

Documentação detalhada de cada módulo funcional e técnico identificado no código-fonte do OptiManager (`1.0.S7 Beta`):

| Módulo | Descrição | Documento |
| :--- | :--- | :--- |
| **Autenticação e Usuários** | Assistente inicial de 6 passos, login HMAC-SHA256, RBAC e gestão de usuários | [autenticacao-e-usuarios.md](./autenticacao-e-usuarios.md) |
| **Dashboard e Atalhos** | Painel inicial com métricas, ordens recentes, recomendações e atalhos rápidos | [dashboard-e-atalhos.md](./dashboard-e-atalhos.md) |
| **Ordens de Serviço** | Ciclo completo de OS: criação, edição, lote, entrega e baixa de pagamento | [ordens-de-servico.md](./ordens-de-servico.md) |
| **Clientes** | Diretório de clientes, autocompletar, edição, exclusão em lote e WhatsApp | [clientes.md](./clientes.md) |
| **Financeiro** | Indicadores de faturamento, fluxo de recebimentos, impressão e exportação CSV | [financeiro.md](./financeiro.md) |
| **Relatórios e PDF** | Gerador de documentos monocromáticos (A4) na UI e motor PDF Maroto v2 | [relatorios-e-pdf.md](./relatorios-e-pdf.md) |
| **Configurações** | Dados da empresa, logo, accent colors, presets de OS, auto-logout e Sobre | [configuracoes.md](./configuracoes.md) |
| **Backup e Restauração** | Snapshots atômicos `VACUUM INTO`, `.zip`, agendador, retenção e restauração | [backup-e-restauracao.md](./backup-e-restauracao.md) |
| **Auditoria e Privacidade (LGPD)** | Tabela `audit_log`, redação automática de PII em logs e extração segura de IP | [auditoria-e-privacidade.md](./auditoria-e-privacidade.md) |
| **Migração Legada DBF** | Utilitário CLI `profile-dbf` para inspeção de bases DBF legadas | [migracao-legada-dbf.md](./migracao-legada-dbf.md) |
| **Saúde e Monitoramento** | Endpoint `GET /api/health` para verificação do SQLite e container Docker | [saude-e-monitoramento.md](./saude-e-monitoramento.md) |
