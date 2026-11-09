<div align="center">

# OptiManager

**Self-Hosted Business & Service Order Management System**

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Vue 3](https://img.shields.io/badge/Vue-3.5+-4FC08D?style=flat&logo=vuedotjs)](https://vuejs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.7+-3178C6?style=flat&logo=typescript)](https://www.typescriptlang.org/)
[![SQLite](https://img.shields.io/badge/SQLite-WAL%20Mode-003B57?style=flat&logo=sqlite)](https://www.sqlite.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![PWA](https://img.shields.io/badge/PWA-Installable-5A0FC8?style=flat&logo=pwa)](https://developer.mozilla.org/en-US/docs/Web/Progressive_web_apps)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)
[![Status: Early Development](https://img.shields.io/badge/Status-Early%20Development-orange.svg)](#project-status)

*A lightweight, fast, self-contained web application designed for micro and small business owners, technical freelancers, and service providers to manage work orders, client directories, technical reports, and billing with full data sovereignty.*

**Official Website:** [https://jvfigueiro.com/projects/optimanager](https://jvfigueiro.com/projects/optimanager)

</div>

---

## Project Status

> [!NOTE]
> **Early Stage Development:**
> OptiManager is currently in an early stage of development. Features, internal APIs, database schemas, and workflows are actively evolving and subject to change.

---

## Key Features

- **Guided First-Run Wizard:** Smooth 6-step initial onboarding wizard (welcome screen, terms of use, choice between fresh setup or database restore, administrator account creation, and business information).
- **End-to-End Service Order (OS) Management:** Create, track, and filter work orders across lifecycle stages (Quote, Open, In Progress, Completed, Delivered, Cancelled) with assignee, deadlines, and financial breakdown.
- **Client Directory with Smart Autocomplete:** Maintain customer records for instant lookup and automated autofill across service orders and PDF reports.
- **Professional Reports & Receipts:** Generate and print clean, client-ready Technical Reports, Quotes, and Delivery Receipts.
- **Financial Dashboard & Metrics:** Real-time visibility into monthly revenue, daily turnover, average ticket, pending balances, and CSV export.
- **Progressive Web App (PWA):** Fully installable on Desktop (Chrome, Edge) and mobile devices (Android, iOS) with a service worker for rapid offline shell loading.
- **Security & Access Control:** Strict JWT-based authentication with boot-time HMAC secret entropy checks, role-based access control (RBAC), and persistent audit logging.
- **Transactional Backup Engine:** Consistent atomic snapshots via SQLite native `VACUUM INTO`, pre-execution integrity checks (`PRAGMA integrity_check`), segregated retention policies, and automatic pre-migration safeguards.
- **Modern Responsive Interface:** Clean UI with support for both Light and Dark themes.

---

## Architecture & Technology Stack

OptiManager is engineered for **minimal footprint** (< 40MB operational RAM), zero heavy runtime dependencies, and effortless deployment:

- **Backend:** [Go](https://go.dev/) (Go 1.24+), utilizing the standard library `http.ServeMux` and lightweight middleware for CORS, gzip compression, logging, and panic recovery.
- **Database:** Embedded [SQLite](https://www.sqlite.org/) configured in WAL mode (*Write-Ahead Logging*), dedicated connection pools (1 writer, 8 readers), numbered migrations, and schema tracking via `PRAGMA user_version`.
- **Frontend:** [Vue 3](https://vuejs.org/) (Composition API), [TypeScript](https://www.typescriptlang.org/), [Vite](https://vitejs.dev/), [Tailwind CSS](https://tailwindcss.com/), [Pinia](https://pinia.vuejs.org/), and [Lucide Icons](https://lucide.dev/).
- **Containerization:** Multi-stage [Docker](https://www.docker.com/) build yielding a minimal Alpine runtime image where all compiled frontend assets are directly embedded into the single Go binary (`//go:embed`).

---

## Quick Start with Docker (Recommended)

### Prerequisites
- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/) installed on your machine.

### 1. Clone the Repository
```bash
git clone https://github.com/jvfigueiro/optimanager.git
cd optimanager
```

### 2. Initialize the Environment File (.env)
The repository provides a utility script that generates a cryptographically secure HMAC secret without overwriting existing configuration files:

```bash
chmod +x scripts/init-env.sh
./scripts/init-env.sh
```

> **Note:** This creates `.env` with a 64-character hexadecimal `JWT_SECRET` generated via `openssl rand -hex 32`.

### 3. Launch the Application Container
```bash
docker compose up -d
```

### 4. Open in Your Browser
Navigate to:
```
http://localhost:8585
```
Follow the **Welcome Wizard** to configure your administrator user and business details.

---

## Environment Variables

The `.env` file governs runtime configuration:

| Variable | Description | Default | Required in Production |
| :--- | :--- | :--- | :--- |
| `PORT` | Public HTTP listening port | `8585` | No |
| `DB_PATH` | Path to the SQLite database file inside container | `/data/optimanager.db` | No |
| `TZ` | Container and application time zone | `America/Sao_Paulo` | No |
| `JWT_SECRET` | 256-bit HMAC signing key (minimum 32 characters) | *Generated by script* | **Yes** |
| `APP_ENV` | Application environment (`production` or `dev`) | `production` | No |
| `TRUST_PROXY` | When `true`, honors `X-Forwarded-For` for client IP logging | `false` | No |

---

## Local Development

To run or develop outside of Docker:

### Backend (Go)
```bash
# Run unit, integration, and race-condition tests
go test ./... -race -count=1

# Run static analysis
go vet ./...

# Start local development server
APP_ENV=dev DB_PATH=./data/optimanager.db go run cmd/server/main.go
```

### Frontend (Vue 3 / Vite)
```bash
cd ui
npm install
npm run build   # Type-check (vue-tsc) and production build
npm run dev     # Vite development server with HMR
```

---

## Directory Structure

```
optimanager/
├── cmd/
│   └── server/             # Application entry point (main.go)
├── docs/                   # Technical specifications and architectural plans
├── internal/
│   ├── database/           # SQLite connection pool, migrations, and repositories
│   │   └── migrations/     # Versioned SQL scripts tracked by PRAGMA user_version
│   ├── handlers/           # REST HTTP endpoints (Auth, OS, Clients, Config, Audit)
│   ├── models/             # Domain entities and request/response DTOs
│   └── services/
│       ├── backup/         # Atomic backup engine (VACUUM INTO, retention, integrity)
│       └── pdf/            # Document and report generation utilities
├── scripts/
│   └── init-env.sh         # Helper script for secure .env key provisioning
├── ui/                     # Single Page Application frontend (Vue 3 + Vite)
│   ├── public/             # Static assets, web manifest, and Service Worker
│   └── src/
│       ├── components/     # Modals, layout wrappers, and reusable UI components
│       ├── stores/         # Global reactive state management (Pinia)
│       └── views/          # Primary views (Dashboard, Orders, Clients, Finance, etc.)
├── Dockerfile              # Multi-stage production container build (Node + Go -> Alpine)
├── docker-compose.yml      # Container orchestration with persistent storage volume
└── .env.example            # Environment variables template
```

---

## License

This software is licensed under the terms of the **GNU General Public License v3.0 (GPLv3)**. See the [LICENSE](LICENSE) file for complete details.
