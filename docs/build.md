# Guia de Compilação, Empacotamento e Distribuição

O OptiManager compila o frontend Vue 3 em arquivos estáticos (`ui/dist/`) e os embute diretamente no binário Go (`//go:embed all:dist`), produzindo um único executável sem dependências externas em tempo de execução (`CGO_ENABLED=0`).

## Pré-requisitos de Build
- **Node.js** (testado com `node:22-alpine` no `Dockerfile`) + `npm`
- **Go** `1.26+` (conforme `go.mod`)
- **Docker** e **Docker Compose** (para build em container)

---

## 1. Compilação via Makefile

O `Makefile` na raiz automatiza todos os alvos principais:

```bash
# Compilar apenas a interface web (type-check vue-tsc + Vite)
make build-ui

# Compilar binário Linux amd64 em dist/optimanager-linux-amd64
make build-linux

# Compilar binário portátil Windows amd64 + copiar script PowerShell para dist/
make build-windows

# Executar análise estática e testes Go com detector de race
make test

# Limpar pasta dist/
make clean
```

Você também pode usar o script dedicado para Windows:
```bash
./scripts/build-windows.sh
```
Artefatos gerados em `dist/`:
- `dist/optimanager-windows-amd64.exe`
- `dist/setup-optimanager.ps1` (copiado por `make build-windows`)

> **Nota sobre instalador:** O projeto não utiliza instalador MSI/NSIS/InnoSetup. A distribuição Windows é portátil (copiar o `.exe` e, opcionalmente, rodar o `setup-optimanager.ps1` como Administrador para liberar a porta no Firewall e registrar a Tarefa Agendada). Consulte [WINDOWS-PORTABLE.md](./WINDOWS-PORTABLE.md).

---

## 2. Empacotamento Docker

O `Dockerfile` executa um build multi-stage em 3 etapas:
1. `frontend-builder` (`node:22-alpine`): `npm ci && npm run build` em `ui/`.
2. `backend-builder` (`golang:alpine`): copia `ui/dist` e compila `CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o optimanager ./cmd/server`.
3. Imagem final (`alpine:3.21`): instala `ca-certificates` e `tzdata` (`TZ=America/Sao_Paulo`), expõe a porta `8585` e monta o volume `/data`.

Para construir e iniciar com Docker Compose:
```bash
./scripts/init-env.sh
docker compose up -d --build
```
