package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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
		t.Fatalf("failed to initialize test database: %v", err)
	}

	r := chi.NewRouter()

	t.Setenv("JWT_SECRET", "test-secret-at-least-32-characters-long-1234567890")

	healthH := handlers.NewHealthHandler(db)
	authH, err := handlers.NewAuthHandler(db)
	if err != nil {
		t.Fatalf("failed to initialize auth handler: %v", err)
	}
	dashH := handlers.NewDashboardHandler(db)
	cliH := handlers.NewClienteHandler(db)
	osH := handlers.NewOrdemHandler(db)
	cfgH := handlers.NewConfigHandler(db)
	atH := handlers.NewAtalhoHandler(db)

	r.Route("/api", func(api chi.Router) {
		api.Get("/health", healthH.Check)
		api.Get("/config", cfgH.Get)
		api.Route("/auth", func(ar chi.Router) {
			ar.Get("/status", authH.Status)
			ar.Post("/setup", authH.Setup)
			ar.Post("/login", authH.Login)
			ar.Post("/logout", authH.Logout)
		})

		api.Group(func(protected chi.Router) {
			protected.Use(authH.RequireAuth)

			protected.Get("/dashboard", dashH.Get)
			protected.Put("/config", cfgH.Update)

			protected.Route("/usuarios", func(ur chi.Router) {
				ur.Get("/", authH.ListUsuarios)
				ur.Post("/", authH.CreateUsuario)
				ur.Put("/{id}", authH.UpdateUsuario)
				ur.Delete("/{id}", authH.DeleteUsuario)
			})
			protected.Route("/clientes", func(cr chi.Router) {
				cr.Get("/", cliH.List)
				cr.Post("/", cliH.Create)
				cr.Post("/batch", cliH.Batch)
				cr.Delete("/{id}", cliH.Delete)
			})
			protected.Route("/ordens", func(or chi.Router) {
				or.Get("/", osH.List)
				or.Post("/", osH.Create)
				or.Post("/batch", osH.Batch)
				or.Put("/{id}", osH.Update)
				or.Delete("/{id}", osH.Delete)
				or.Patch("/{id}/baixa", osH.BaixaPagamento)
			})
			protected.Route("/atalhos", func(ar chi.Router) {
				ar.Get("/", atH.List)
				ar.Post("/", atH.Create)
				ar.Put("/{id}", atH.Update)
				ar.Delete("/{id}", atH.Delete)
			})
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

func setupAdminAndGetToken(t *testing.T, r *chi.Mux) string {
	t.Helper()
	setupPayload := models.SetupAdminRequest{
		Nome:         "Administrador de Teste",
		Email:        "admin@empresa.com.br",
		Senha:        "password123",
		Fantasia:     "Tech Soluções de Teste",
		RazaoSocial:  "Tech Soluções & Serviços Ltda",
		CNPJ:         "00.000.000/0001-00",
		Fones:        "(11) 98765-4321",
		EmailContato: "contato@empresa.com.br",
		Endereco:     "Av. Paulista, 1000 - Bela Vista",
		Cidade:       "São Paulo - SP",
	}
	body, _ := json.Marshal(setupPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/setup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 on admin setup, got %d: %s", w.Code, w.Body.String())
	}

	var authResp models.AuthResponse
	if err := json.NewDecoder(w.Body).Decode(&authResp); err != nil {
		t.Fatalf("failed to decode setup response: %v", err)
	}
	return authResp.Token
}

func TestHealthCheck(t *testing.T) {
	r, _, teardown := setupTestServer(t)
	defer teardown()

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp handlers.HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Database != "connected" || resp.JournalMode != "wal" {
		t.Fatalf("expected connected database in WAL mode, got %s / %s", resp.Database, resp.JournalMode)
	}
}

func TestAuthFlowAndUsuariosCRUD(t *testing.T) {
	r, _, teardown := setupTestServer(t)
	defer teardown()

	// 1. Initial status check (needs_setup = true)
	reqStatus := httptest.NewRequest(http.MethodGet, "/api/auth/status", nil)
	wStatus := httptest.NewRecorder()
	r.ServeHTTP(wStatus, reqStatus)

	var st models.AuthStatusResponse
	_ = json.NewDecoder(wStatus.Body).Decode(&st)
	if !st.NeedsSetup || st.Authenticated {
		t.Fatalf("expected needs_setup=true and authenticated=false, got %+v", st)
	}

	// 2. Protected route without token must return 401
	reqUnauth := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	wUnauth := httptest.NewRecorder()
	r.ServeHTTP(wUnauth, reqUnauth)
	if wUnauth.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", wUnauth.Code)
	}

	// 3. Admin setup
	token := setupAdminAndGetToken(t, r)

	// 4. Duplicate setup must be blocked with 403
	setupBody, _ := json.Marshal(models.SetupAdminRequest{
		Nome:  "Outro Admin",
		Email: "outro@optimanager.com",
		Senha: "password123",
	})
	reqDupSetup := httptest.NewRequest(http.MethodPost, "/api/auth/setup", bytes.NewReader(setupBody))
	wDupSetup := httptest.NewRecorder()
	r.ServeHTTP(wDupSetup, reqDupSetup)
	if wDupSetup.Code != http.StatusForbidden {
		t.Fatalf("expected 403 on duplicate setup, got %d", wDupSetup.Code)
	}

	// 5. Create second user via /api/usuarios
	newUsrPayload := models.CreateUsuarioRequest{
		Nome:  "Operador 1",
		Email: "op1@optimanager.com",
		Senha: "senha123",
		Role:  "operador",
	}
	usrBody, _ := json.Marshal(newUsrPayload)
	reqCreateUsr := httptest.NewRequest(http.MethodPost, "/api/usuarios", bytes.NewReader(usrBody))
	reqCreateUsr.Header.Set("Authorization", "Bearer "+token)
	wCreateUsr := httptest.NewRecorder()
	r.ServeHTTP(wCreateUsr, reqCreateUsr)
	if wCreateUsr.Code != http.StatusCreated {
		t.Fatalf("expected 201 on create user, got %d: %s", wCreateUsr.Code, wCreateUsr.Body.String())
	}

	// 6. List users (should have 2)
	reqListUsr := httptest.NewRequest(http.MethodGet, "/api/usuarios", nil)
	reqListUsr.Header.Set("Authorization", "Bearer "+token)
	wListUsr := httptest.NewRecorder()
	r.ServeHTTP(wListUsr, reqListUsr)
	var usuarios []models.Usuario
	_ = json.NewDecoder(wListUsr.Body).Decode(&usuarios)
	if len(usuarios) != 2 {
		t.Fatalf("expected 2 users, got %d", len(usuarios))
	}
}

func TestCreateOrdemAndAutoClient(t *testing.T) {
	r, _, teardown := setupTestServer(t)
	defer teardown()

	token := setupAdminAndGetToken(t, r)
	authHeader := "Bearer " + token

	// 1. Create order with new client and VT shortcut
	payload := models.CreateOrdemRequest{
		ClienteNome:     "Advocacia Santos",
		ClienteDoc:      "12.345.678/0001-90",
		ClienteContato:  "(55) 99111-2233",
		Servico:         "Configuração de rede e impressoras",
		IsVT:            true,
		ValorTotal:      0, // Adopts default 80.00
		Status:          "Concluído",
		PagamentoStatus: "Pendente",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/ordens", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 on create order, got %d: %s", w.Code, w.Body.String())
	}

	var osResp models.OrdemServico
	_ = json.NewDecoder(w.Body).Decode(&osResp)

	if osResp.ID < 1000 {
		t.Fatalf("expected ID >= 1000, got %d", osResp.ID)
	}
	if osResp.Equipamento != "N/A" {
		t.Fatalf("expected Equipamento 'N/A', got '%s'", osResp.Equipamento)
	}
	if osResp.ValorTotal != 80.00 {
		t.Fatalf("expected default VT value 80.00, got %.2f", osResp.ValorTotal)
	}
	if osResp.ClienteNome != "Advocacia Santos" {
		t.Fatalf("expected ClienteNome 'Advocacia Santos', got '%s'", osResp.ClienteNome)
	}

	// 2. Verify auto-created client in database
	reqCli := httptest.NewRequest(http.MethodGet, "/api/clientes?search=Advocacia", nil)
	reqCli.Header.Set("Authorization", authHeader)
	wCli := httptest.NewRecorder()
	r.ServeHTTP(wCli, reqCli)

	var clientes []models.Cliente
	_ = json.NewDecoder(wCli.Body).Decode(&clientes)
	if len(clientes) != 1 || clientes[0].Nome != "Advocacia Santos" {
		t.Fatalf("expected 1 auto-created client, got: %v", clientes)
	}

	// 3. Verify dashboard pending metrics
	reqDash := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	reqDash.Header.Set("Authorization", authHeader)
	wDash := httptest.NewRecorder()
	r.ServeHTTP(wDash, reqDash)

	var dash models.DashboardData
	_ = json.NewDecoder(wDash.Body).Decode(&dash)
	if dash.Metricas.PendenteReceber != 80.00 || dash.Metricas.QtdPendentes != 1 {
		t.Fatalf("expected pending 80.00 / 1 order, got: %+v", dash.Metricas)
	}

	// 4. Quick payment settlement
	baixaPayload := models.BaixaPagamentoRequest{Forma: "Pix", Status: "Pago"}
	bBody, _ := json.Marshal(baixaPayload)
	reqBaixa := httptest.NewRequest(http.MethodPatch, "/api/ordens/1000/baixa", bytes.NewReader(bBody))
	reqBaixa.Header.Set("Authorization", authHeader)
	wBaixa := httptest.NewRecorder()
	r.ServeHTTP(wBaixa, reqBaixa)

	if wBaixa.Code != http.StatusOK {
		t.Fatalf("expected status 200 on payment settlement, got %d: %s", wBaixa.Code, wBaixa.Body.String())
	}

	// 5. Verify dashboard after payment settlement
	wDash2 := httptest.NewRecorder()
	reqDash2 := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	reqDash2.Header.Set("Authorization", authHeader)
	r.ServeHTTP(wDash2, reqDash2)

	var dash2 models.DashboardData
	_ = json.NewDecoder(wDash2.Body).Decode(&dash2)
	if dash2.Metricas.CaixaRecebido != 80.00 || dash2.Metricas.PendenteReceber != 0 || dash2.Metricas.QtdPendentes != 0 {
		t.Fatalf("expected received 80.00 and pending 0, got: %+v", dash2.Metricas)
	}

	// 6. Delete order
	reqDel := httptest.NewRequest(http.MethodDelete, "/api/ordens/1000", nil)
	reqDel.Header.Set("Authorization", authHeader)
	wDel := httptest.NewRecorder()
	r.ServeHTTP(wDel, reqDel)
	if wDel.Code != http.StatusOK {
		t.Fatalf("expected status 200 on order delete, got %d: %s", wDel.Code, wDel.Body.String())
	}

	// 7. Test Batch Operations on Orders
	// Create two new orders
	for i := 1; i <= 2; i++ {
		oPayload := fmt.Sprintf(`{"servico":"Batch Service %d","cliente_nome":"Batch Client %d","valor_total":100.0}`, i, i)
		wCreate := httptest.NewRecorder()
		reqCreate := httptest.NewRequest(http.MethodPost, "/api/ordens", strings.NewReader(oPayload))
		reqCreate.Header.Set("Authorization", authHeader)
		reqCreate.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(wCreate, reqCreate)
		if wCreate.Code != http.StatusCreated {
			t.Fatalf("failed to create order for batch test: %s", wCreate.Body.String())
		}
	}

	// Batch Status Update
	wBatchStatus := httptest.NewRecorder()
	reqBatchStatus := httptest.NewRequest(http.MethodPost, "/api/ordens/batch", strings.NewReader(`{"action":"status","ids":[1001,1002],"status":"Em Serviço"}`))
	reqBatchStatus.Header.Set("Authorization", authHeader)
	reqBatchStatus.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(wBatchStatus, reqBatchStatus)
	if wBatchStatus.Code != http.StatusOK {
		t.Fatalf("batch status failed: %s", wBatchStatus.Body.String())
	}

	// Batch Baixa Update
	wBatchBaixa := httptest.NewRecorder()
	reqBatchBaixa := httptest.NewRequest(http.MethodPost, "/api/ordens/batch", strings.NewReader(`{"action":"baixa","ids":[1001,1002],"pagamento_forma":"Pix"}`))
	reqBatchBaixa.Header.Set("Authorization", authHeader)
	reqBatchBaixa.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(wBatchBaixa, reqBatchBaixa)
	if wBatchBaixa.Code != http.StatusOK {
		t.Fatalf("batch baixa failed: %s", wBatchBaixa.Body.String())
	}

	// Batch Duplicar
	wBatchDup := httptest.NewRecorder()
	reqBatchDup := httptest.NewRequest(http.MethodPost, "/api/ordens/batch", strings.NewReader(`{"action":"duplicar","ids":[1001]}`))
	reqBatchDup.Header.Set("Authorization", authHeader)
	reqBatchDup.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(wBatchDup, reqBatchDup)
	if wBatchDup.Code != http.StatusOK {
		t.Fatalf("batch duplicate failed: %s", wBatchDup.Body.String())
	}

	// Batch Delete Orders
	wBatchDel := httptest.NewRecorder()
	reqBatchDel := httptest.NewRequest(http.MethodPost, "/api/ordens/batch", strings.NewReader(`{"action":"delete","ids":[1001,1002,1003]}`))
	reqBatchDel.Header.Set("Authorization", authHeader)
	reqBatchDel.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(wBatchDel, reqBatchDel)
	if wBatchDel.Code != http.StatusOK {
		t.Fatalf("batch delete failed: %s", wBatchDel.Body.String())
	}

	// 8. Test Batch Operations on Clients
	// Create free clients (not linked to OS)
	wCliCreate := httptest.NewRecorder()
	reqCliCreate := httptest.NewRequest(http.MethodPost, "/api/clientes", strings.NewReader(`{"nome":"Cliente Solto Teste"}`))
	reqCliCreate.Header.Set("Authorization", authHeader)
	reqCliCreate.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(wCliCreate, reqCliCreate)
	var cliSolto models.Cliente
	_ = json.NewDecoder(wCliCreate.Body).Decode(&cliSolto)

	wBatchCli := httptest.NewRecorder()
	reqBatchCli := httptest.NewRequest(http.MethodPost, "/api/clientes/batch", strings.NewReader(fmt.Sprintf(`{"action":"delete","ids":[%d]}`, cliSolto.ID)))
	reqBatchCli.Header.Set("Authorization", authHeader)
	reqBatchCli.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(wBatchCli, reqBatchCli)
	if wBatchCli.Code != http.StatusOK {
		t.Fatalf("batch client delete failed: %s", wBatchCli.Body.String())
	}
}
