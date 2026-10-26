package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"

	"optimanager/internal/database"
	"optimanager/internal/handlers"
	"optimanager/internal/models"
)

func setupTestServer(t *testing.T) (*chi.Mux, *database.DB, func()) {
	t.Helper()

	testDBPath := "./test_unit_opti.db"
	_ = os.Remove(testDBPath)
	_ = os.Remove(testDBPath + "-wal")
	_ = os.Remove(testDBPath + "-shm")

	db, err := database.New(testDBPath)
	if err != nil {
		t.Fatalf("Erro ao inicializar banco de teste: %v", err)
	}

	r := chi.NewRouter()

	healthH := handlers.NewHealthHandler(db)
	dashH := handlers.NewDashboardHandler(db)
	cliH := handlers.NewClienteHandler(db)
	osH := handlers.NewOrdemHandler(db)
	cfgH := handlers.NewConfigHandler(db)
	atH := handlers.NewAtalhoHandler(db)

	r.Route("/api", func(api chi.Router) {
		api.Get("/health", healthH.Check)
		api.Get("/dashboard", dashH.Get)
		api.Route("/clientes", func(cr chi.Router) {
			cr.Get("/", cliH.List)
			cr.Post("/", cliH.Create)
		})
		api.Route("/ordens", func(or chi.Router) {
			or.Get("/", osH.List)
			or.Post("/", osH.Create)
			or.Patch("/{id}/baixa", osH.BaixaPagamento)
		})
		api.Route("/config", func(cfr chi.Router) {
			cfr.Get("/", cfgH.Get)
		})
		api.Route("/atalhos", func(ar chi.Router) {
			ar.Get("/", atH.List)
			ar.Post("/", atH.Create)
		})
	})

	teardown := func() {
		db.Close()
		_ = os.Remove(testDBPath)
		_ = os.Remove(testDBPath + "-wal")
		_ = os.Remove(testDBPath + "-shm")
	}

	return r, db, teardown
}

func TestHealthCheck(t *testing.T) {
	r, _, teardown := setupTestServer(t)
	defer teardown()

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperado status 200, recebido %d", w.Code)
	}

	var resp handlers.HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("erro ao decodificar resposta: %v", err)
	}
	if resp.Database != "connected" || resp.JournalMode != "wal" {
		t.Fatalf("esperado banco conectado em wal, recebido %s / %s", resp.Database, resp.JournalMode)
	}
}

func TestCreateOrdemAndAutoClient(t *testing.T) {
	r, _, teardown := setupTestServer(t)
	defer teardown()

	// 1. Cria OS com cliente novo e atalho VT
	payload := models.CreateOrdemRequest{
		ClienteNome:     "Advocacia Santos",
		ClienteDoc:      "12.345.678/0001-90",
		ClienteContato:  "(55) 99111-2233",
		Servico:         "Configuração de rede e impressoras",
		IsVT:            true,
		ValorTotal:      0, // Deve adotar 80.00
		Status:          "Concluído",
		PagamentoStatus: "Pendente",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/ordens", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("esperado status 201 ao criar OS, recebido %d: %s", w.Code, w.Body.String())
	}

	var osResp models.OrdemServico
	_ = json.NewDecoder(w.Body).Decode(&osResp)

	if osResp.ID < 1000 {
		t.Fatalf("esperado ID >= 1000, recebido %d", osResp.ID)
	}
	if osResp.Equipamento != "N/A" {
		t.Fatalf("esperado Equipamento 'N/A', recebido '%s'", osResp.Equipamento)
	}
	if osResp.ValorTotal != 80.00 {
		t.Fatalf("esperado valor padrão VT 80.00, recebido %.2f", osResp.ValorTotal)
	}
	if osResp.ClienteNome != "Advocacia Santos" {
		t.Fatalf("esperado ClienteNome 'Advocacia Santos', recebido '%s'", osResp.ClienteNome)
	}

	// 2. Verifica se cliente foi cadastrado na base
	reqCli := httptest.NewRequest(http.MethodGet, "/api/clientes?search=Advocacia", nil)
	wCli := httptest.NewRecorder()
	r.ServeHTTP(wCli, reqCli)

	var clientes []models.Cliente
	_ = json.NewDecoder(wCli.Body).Decode(&clientes)
	if len(clientes) != 1 || clientes[0].Nome != "Advocacia Santos" {
		t.Fatalf("esperado 1 cliente cadastrado automaticamente, recebido: %v", clientes)
	}

	// 3. Verifica Dashboard com a pendência
	reqDash := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	wDash := httptest.NewRecorder()
	r.ServeHTTP(wDash, reqDash)

	var dash models.DashboardData
	_ = json.NewDecoder(wDash.Body).Decode(&dash)
	if dash.Metricas.PendenteReceber != 80.00 || dash.Metricas.QtdPendentes != 1 {
		t.Fatalf("esperado pendente 80.00 / 1 pendência, recebido: %+v", dash.Metricas)
	}

	// 4. Baixa de pagamento
	baixaPayload := models.BaixaPagamentoRequest{Forma: "Pix", Status: "Pago"}
	bBody, _ := json.Marshal(baixaPayload)
	reqBaixa := httptest.NewRequest(http.MethodPatch, "/api/ordens/1000/baixa", bytes.NewReader(bBody))
	wBaixa := httptest.NewRecorder()
	r.ServeHTTP(wBaixa, reqBaixa)

	if wBaixa.Code != http.StatusOK {
		t.Fatalf("esperado status 200 na baixa, recebido %d: %s", wBaixa.Code, wBaixa.Body.String())
	}

	// 5. Verifica Dashboard após baixa
	wDash2 := httptest.NewRecorder()
	r.ServeHTTP(wDash2, httptest.NewRequest(http.MethodGet, "/api/dashboard", nil))

	var dash2 models.DashboardData
	_ = json.NewDecoder(wDash2.Body).Decode(&dash2)
	if dash2.Metricas.CaixaRecebido != 80.00 || dash2.Metricas.PendenteReceber != 0 || dash2.Metricas.QtdPendentes != 0 {
		t.Fatalf("esperado caixa recebido 80.00 e pendente 0, recebido: %+v", dash2.Metricas)
	}
}

