<div align="center">

# OptiManager

**Sistema Web Auto-Hospedado para Gestão de Ordens de Serviço e Pequenos Negócios**

[![Versão](https://img.shields.io/badge/Vers%C3%A3o-1.0.S7%20Beta-0B5694?style=flat)](docs/arch.md)
[![Plataforma](https://img.shields.io/badge/Plataforma-Linux%20%7C%20Windows%20Port%C3%A1til%20%7C%20Docker-2496ED?style=flat)](docs/build.md)
[![Licença: GPL v3](https://img.shields.io/badge/Licen%C3%A7a-GPLv3-blue.svg)](LICENSE)

*Aplicação web leve, rápida e auto-contida em binário único (Go + Vue 3 + SQLite), projetada para micro e pequenos empreendedores, assistências técnicas e prestadores de serviço com soberania total dos dados.*

**Site Oficial:** [https://jvfigueiro.com/projects/optimanager](https://jvfigueiro.com/projects/optimanager)

</div>

---

## Capturas de Tela

| Dashboard | Serviços (Ordens de Serviço) |
| :---: | :---: |
| ![Dashboard do OptiManager](docs/screenshots/dashboard.png) | ![Tela de Serviços](docs/screenshots/servicos.png) |

| Clientes | Gerador de Documentos (A4 Monocromático) |
| :---: | :---: |
| ![Tela de Clientes](docs/screenshots/clientes.png) | ![Gerador de Documentos](docs/screenshots/relatorios.png) |

| Configurações |
| :---: |
| ![Tela de Configurações](docs/screenshots/configuracoes.png) |

---

## Principais Recursos

| Módulo | Descrição | Documentação |
| :--- | :--- | :--- |
| **Autenticação e Usuários** | Assistente inicial de 6 passos, login HMAC-SHA256 e controle RBAC | [autenticacao-e-usuarios.md](docs/modules/autenticacao-e-usuarios.md) |
| **Dashboard e Atalhos** | Métricas resumidas, recomendações, ações frequentes e atalhos rápidos | [dashboard-e-atalhos.md](docs/modules/dashboard-e-atalhos.md) |
| **Ordens de Serviço** | Criação, edição, conclusão, ações em lote e baixa de pagamento | [ordens-de-servico.md](docs/modules/ordens-de-servico.md) |
| **Clientes** | Cadastro com busca, autocompletar em OS e contato via WhatsApp | [clientes.md](docs/modules/clientes.md) |
| **Financeiro** | Fluxo de recebimentos, indicadores e exportação CSV | [financeiro.md](docs/modules/financeiro.md) |
| **Relatórios e PDF** | Gerador A4 monocromático na UI e PDF nativo via Maroto v2 | [relatorios-e-pdf.md](docs/modules/relatorios-e-pdf.md) |
| **Configurações** | Logo, accent colors, presets de OS/Serviços e auto-logout | [configuracoes.md](docs/modules/configuracoes.md) |
| **Backup e Restauração** | Snapshots atômicos `VACUUM INTO`, `.zip`, agendador e retenção | [backup-e-restauracao.md](docs/modules/backup-e-restauracao.md) |
| **Auditoria e Privacidade** | Tabela `audit_log`, redação LGPD em logs/erros e `TRUST_PROXY` | [auditoria-e-privacidade.md](docs/modules/auditoria-e-privacidade.md) |
| **Migração Legada DBF** | Subcomando CLI `profile-dbf` para inspeção de bases DBF | [migracao-legada-dbf.md](docs/modules/migracao-legada-dbf.md) |
| **Saúde e Monitoramento** | Endpoint `GET /api/health` e healthcheck Docker | [saude-e-monitoramento.md](docs/modules/saude-e-monitoramento.md) |

---

## Requisitos

- **Via Docker (Recomendado):** Docker e Docker Compose.
- **Via Binário Portátil (Windows/Linux):** nenhum runtime externo necessário (`CGO_ENABLED=0`).
- **Para compilar do código-fonte:** Go `1.26+` e Node.js `22+` (`npm`).

---

## Instalação e Primeiros Passos

### Opção 1: Docker Compose
```bash
git clone https://github.com/jvfigueiro/optimanager.git
cd optimanager
./scripts/init-env.sh
docker compose up -d
```
Acesse `http://localhost:8585` e siga o **Assistente de Boas-Vindas** (6 passos) para configurar o administrador e os dados da empresa.

### Opção 2: Windows Portátil
Gere ou copie `optimanager-windows-amd64.exe` para uma pasta dedicada (ex.: `C:\OptiManager`) e execute-o. Para liberar a porta `8585` no Firewall e configurar inicialização automática com o Windows, consulte [docs/WINDOWS-PORTABLE.md](docs/WINDOWS-PORTABLE.md).

---

## Visão Geral da Arquitetura Modular

O OptiManager é um **monólito modular em binário único**: o frontend SPA é embutido no executável Go via `//go:embed all:dist` e os módulos verticais combinam modelos (`internal/models`), repositórios SQLite WAL (`internal/database`), serviços (`internal/services`), handlers HTTP (`internal/handlers`) e stores/views Vue (`ui/src`).

Detalhes completos e diagramas Mermaid em [docs/arch.md](docs/arch.md).

---

## Tecnologias Utilizadas

- **Backend:** Go `1.26.1`, `chi/v5`, `modernc.org/sqlite` (WAL), `maroto/v2` (PDF), `golang.org/x/crypto` (`bcrypt`), `golang.org/x/text`
- **Frontend:** Vue `3.5`, TypeScript `5.7`, Pinia `2.3`, Tailwind CSS `3.4`, Vite `6.1`, `lucide-vue-next`, PWA
- **Infraestrutura:** Docker multi-stage (`alpine:3.21`), `Makefile`, PowerShell (`setup-optimanager.ps1`)

---

## Estrutura do Projeto

```text
optimanager/
├── cmd/server/          # Ponto de entrada HTTP e CLI `profile-dbf`
├── internal/            # Núcleo Go (database, handlers, models, money, redact, services, legacy/dbf)
├── ui/                  # SPA Vue 3 + TypeScript + embedding `//go:embed all:dist`
├── scripts/             # Scripts de build, ambiente, demo e setup Windows
└── docs/                # Documentação técnica, módulos e guias
```

---

## Como Contribuir e Criar Módulos

Consulte o guia prático baseado no código real em [docs/create-module.md](docs/create-module.md) e as instruções de compilação em [docs/build.md](docs/build.md).

---

## Aviso de Responsabilidade

O OptiManager armazena dados operacionais localmente no banco SQLite e permite restaurar backups que substituem a base ativa. No Windows, a execução opcional do script `scripts/windows/setup-optimanager.ps1` como Administrador altera configurações do sistema (cria regra de entrada no Firewall do Windows e registra uma Tarefa Agendada com privilégios elevados). Mantenha cópias de backup externas periódicas.

---

## Documentação Completa

- [Mapa de Arquitetura (`docs/_analysis/00-map.md`)](docs/_analysis/00-map.md)
- [Visão Geral e Diagramas (`docs/arch.md`)](docs/arch.md)
- [Índice de Módulos (`docs/modules/README.md`)](docs/modules/README.md)
- [Como Criar um Novo Módulo (`docs/create-module.md`)](docs/create-module.md)
- [Compilação e Empacotamento (`docs/build.md`)](docs/build.md)
- [Configurações, Logs e Auditoria (`docs/logs-and-settings.md`)](docs/logs-and-settings.md)
- [FAQ e Solução de Problemas (`docs/faq-and-troubleshooting.md`)](docs/faq-and-troubleshooting.md)
- [Guia Windows Portátil (`docs/WINDOWS-PORTABLE.md`)](docs/WINDOWS-PORTABLE.md)
- [Dados, LGPD e Backups (`docs/DADOS-E-BACKUPS.md`)](docs/DADOS-E-BACKUPS.md)

---

## Licença

Distribuído sob a licença **GNU General Public License v3.0 (GPLv3)**. Consulte o arquivo [LICENSE](LICENSE).
