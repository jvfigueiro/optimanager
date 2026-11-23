# Como Criar um Novo Módulo no OptiManager

O OptiManager segue uma arquitetura **modular em camadas verticais** compiladas estaticamente (não há carregamento dinâmico de plugins em runtime). Para adicionar um novo recurso completo (ex.: `lembretes`), siga o padrão usado pelos módulos existentes como `atalho`:

## 1. Modelo de Domínio (`internal/models/`)
Crie `internal/models/lembrete.go` com a struct da entidade e payloads de entrada:

```go
package models

import "time"

type Lembrete struct {
	ID        int64     `json:"id"`
	Titulo    string    `json:"titulo"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateLembreteRequest struct {
	Titulo string `json:"titulo"`
}
```

## 2. Migração SQL e Repositório (`internal/database/`)
1. Adicione uma nova migration numerada em `internal/database/migrations/005_lembretes.sql` e registre-a no slice `migrations` em `internal/database/db.go` (via `//go:embed`).
2. Crie `internal/database/lembrete_repo.go` com métodos em `*DB`, usando `d.Reader` para `SELECT` e `d.Writer` para `INSERT`/`UPDATE`/`DELETE`:

```go
package database

import "optimanager/internal/models"

func (d *DB) ListLembretes() ([]models.Lembrete, error) {
	rows, err := d.Reader.Query(`SELECT id, titulo, created_at FROM lembretes ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Lembrete
	for rows.Next() {
		var item models.Lembrete
		if err := rows.Scan(&item.ID, &item.Titulo, &item.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, rows.Err()
}
```

## 3. Handler HTTP (`internal/handlers/`)
Crie `internal/handlers/lembrete.go` usando os helpers `respondJSON` e `respondError` (que aplica sanitização LGPD):

```go
package handlers

import (
	"net/http"
	"optimanager/internal/database"
)

type LembreteHandler struct {
	db *database.DB
}

func NewLembreteHandler(db *database.DB) *LembreteHandler {
	return &LembreteHandler{db: db}
}

func (h *LembreteHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.db.ListLembretes()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, items)
}
```

## 4. Registro no Roteador (`cmd/server/main.go`)
Instancie o handler e registre a rota dentro do grupo protegido (`priv` ou `admin`):

```go
lembreteHandler := handlers.NewLembreteHandler(db)
// dentro de priv.Route:
priv.Get("/lembretes", lembreteHandler.List)
```

## 5. Integração no Frontend (`ui/src/`)
1. Adicione as interfaces TypeScript em `ui/src/types/index.ts`.
2. Adicione os métodos de API em `ui/src/api/index.ts` usando `request<T>(...)`.
3. Crie uma store Pinia em `ui/src/stores/lembretes.ts` (opcional, seguindo `stores/clientes.ts`).
4. Para uma nova tela inteira:
   - Crie `ui/src/views/LembretesView.vue`.
   - Adicione o slug ao tipo `ViewType` e `initFromHash()` em `ui/src/stores/navigation.ts`.
   - Adicione o item de menu e renderize a view em `ui/src/components/AppLayout.vue` (e o título em `viewTitles` de `ui/src/App.vue`).

## 6. Validação
Execute os testes e o build completo:
```bash
npm --prefix ui run build
go test ./... -race -count=1
```
