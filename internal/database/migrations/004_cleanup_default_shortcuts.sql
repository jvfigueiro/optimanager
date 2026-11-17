-- Migration 004: Keep only Portal da Nota Fiscal (NFS-e) among legacy default shortcuts

DELETE FROM atalhos
WHERE titulo IN ('PGMEI - Guia DAS', 'Portal do Empreendedor', 'Consulta Cartão CNPJ')
   OR descricao IN ('Gerador de DAS do MEI', 'Serviços e Declaração Anual (DASN)', 'Comprovante de Inscrição Receita');
