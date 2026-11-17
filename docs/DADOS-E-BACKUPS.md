# Política de Dados, Privacidade (LGPD) e Backups no OptiManager

## 1. Minimização de Dados (LGPD)
O OptiManager adota o princípio de minimização de dados para cadastros de clientes e fornecedores:
- **Dados coletados:** apenas `nome` (obrigatório), `contato` (telefone normalizado), `cpf_cnpj` (opcional, validado e único quando preenchido) e `informacoes_adicionais` (texto livre limitado a 500 caracteres).
- **Aviso operacional:** o campo `informacoes_adicionais` e as observações de Ordens de Serviço não devem ser utilizados para registrar dados pessoais sensíveis.
- **Exibição e mascaramento:** o CPF/CNPJ permanece mascarado nas listagens, buscas, relatórios e exportações CSV. A visualização integral na ficha depende da permissão `ver_dados_pessoais` e cada revelação é registrada no `audit_log`.
- **Logs de aplicação:** mensagens de erro e logs do servidor redigem automaticamente padrões de CPF, CNPJ e telefone.

## 2. Conteúdo dos Backups
Os arquivos de backup gerados pelo OptiManager (`.zip` ou `.db` de pré-migração) contêm:
1. O snapshot íntegro do banco de dados SQLite (`optimanager.db`), gerado via `VACUUM INTO` e verificado com `PRAGMA integrity_check`, incluindo todas as tabelas operacionais (`config`, `clientes`, `ordens_servico`, `usuarios`, `atalhos`, `audit_log` e tabelas de catálogo/financeiro).
2. O logotipo configurado pela empresa (`company_logo.*`).
3. Os anexos e fotos de equipamentos armazenados em `data/uploads/` (quando existentes).

## 3. Rotação e Retenção Configurável
O administrador define na tela de Configurações (seção Backups) a quantidade máxima de arquivos mantidos para cada categoria:
- **Automáticos (`auto`):** padrão 7 arquivos.
- **Manuais (`manual`):** padrão 5 arquivos.
- **Pré-restauração (`prerestore`):** padrão 3 arquivos.
- **Pré-migração (`premigration`):** padrão 3 arquivos.

Quando o limite de uma categoria é atingido, os arquivos mais antigos daquela categoria são removidos automaticamente após a confirmação de integridade do novo backup.

## 4. Anonimização de Clientes e Backups Históricos
A ação administrativa **Anonimizar Cliente** substitui irreversivelmente o nome do titular por `Cliente anonimizado #<id>` e limpa `contato`, `cpf_cnpj` e `informacoes_adicionais` no banco de dados ativo, preservando os registros contábeis e de Ordens de Serviço.

**Importante:** a anonimização altera exclusivamente o banco de dados em produção e os próximos backups criados a partir daquele momento. Arquivos de backup `.zip`/`.db` gerados anteriormente são snapshots imutáveis e **não** são reescritos; eles deixarão de existir conforme a janela de retenção configurada (ou podem ser removidos manualmente pelo administrador na tela de Backups caso haja solicitação imediata de descarte).
