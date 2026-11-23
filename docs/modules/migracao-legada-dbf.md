# Módulo: Migração Legada DBF

## Visão geral e problema que resolve
Fornece um utilitário de linha de comando (`profile-dbf`) embutido no próprio binário do OptiManager para inspecionar e gerar perfil estrutural de arquivos `.dbf` (dBase/FoxPro) de sistemas legados antes de planejar migrações de dados.

## Como usar (passo a passo do ponto de vista do técnico)
1. No terminal (Linux ou Windows), execute o binário do OptiManager passando o subcomando `profile-dbf` e o caminho da pasta com arquivos `.dbf`:
   ```bash
   ./optimanager profile-dbf /caminho/para/dbfs
   ```
   Ou no Windows:
   ```cmd
   optimanager-windows-amd64.exe profile-dbf C:\dados\legado
   ```
2. Opções disponíveis:
   - `--encoding cp1252|cp850|cp437|utf-8`: força a codificação de caracteres na leitura.
   - `--json`: imprime o relatório de perfil em formato JSON no `stdout`.

## Como funciona por dentro
1. Em `cmd/server/main.go`, antes de inicializar o banco SQLite ou o servidor HTTP, `main()` verifica `os.Args[1] == "profile-dbf"` e chama `runProfileDBF(os.Args[2:])`.
2. `dbf.ProfileDirectory` lê os cabeçalhos e registros dos arquivos `.dbf` no diretório informado utilizando `golang.org/x/text` para decodificação de codepages DOS/Windows.
3. `dbf.WriteReport` escreve o relatório sumarizado (em texto ou JSON) para `os.Stdout` e encerra o processo sem subir o servidor web.

## Entradas, saídas e configurações
- **Entrada:** diretório contendo arquivos `.dbf` e flags opcionais `--encoding` e `--json`.
- **Saída:** relatório estrutural impresso em `stdout` (não modifica os arquivos `.dbf` nem o SQLite).

## Requisitos
- **Admin no sistema / Windows:** não necessário (apenas permissão de leitura na pasta dos `.dbf`).
- **Dependências externas:** `golang.org/x/text`.

## Riscos e efeitos colaterais
- Operação estritamente de leitura; não altera arquivos `.dbf` originais nem cria banco SQLite.

## Logs e tratamento de erros
- Erros de uso ou leitura são impressos em `os.Stderr` com prefixo `[profile-dbf] Error:` e código de saída `1`.

## Arquivos-fonte relacionados
- `cmd/server/main.go` (`runProfileDBF`)
- `internal/legacy/dbf/dbf.go`
- `internal/legacy/dbf/profiler.go`
- `internal/legacy/dbf/dbf_test.go`
