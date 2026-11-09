package main

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"optimanager/internal/database"
	"optimanager/internal/handlers"
	"optimanager/internal/services/backup"
	"optimanager/ui"
)

func main() {
	log.Println("[OptiManager] Initializing server...")

	// 1. Initialize SQLite database connection and migrations
	db, err := database.New("")
	if err != nil {
		log.Fatalf("[OptiManager] Fatal error initializing database: %v", err)
	}
	defer db.Close()

	// 2. Configure lightweight chi router
	r := chi.NewRouter()

	// Lightweight standard middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(120 * time.Second))

	// CORS middleware for local dev or production
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			origin := req.Header.Get("Origin")
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if req.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, req)
		})
	})

	// 3. Initialize services and HTTP handlers
	backupService, err := backup.NewService(db, "", "")
	if err != nil {
		log.Fatalf("[OptiManager] Fatal error initializing backup service: %v", err)
	}
	schedulerCtx, cancelScheduler := context.WithCancel(context.Background())
	defer cancelScheduler()
	backupService.StartScheduler(schedulerCtx)

	healthHandler := handlers.NewHealthHandler(db)
	authHandler, err := handlers.NewAuthHandler(db)
	if err != nil {
		log.Fatalf("[OptiManager] Fatal authentication initialization: %v", err)
	}
	dashboardHandler := handlers.NewDashboardHandler(db)
	clienteHandler := handlers.NewClienteHandler(db)
	ordemHandler := handlers.NewOrdemHandler(db)
	configHandler := handlers.NewConfigHandler(db)
	atalhoHandler := handlers.NewAtalhoHandler(db)
	backupHandler := handlers.NewBackupHandler(db, backupService)

	// 4. Register API routes
	r.Route("/api", func(api chi.Router) {
		// Public routes
		api.Get("/health", healthHandler.Check)
		api.Get("/config", configHandler.Get)

		api.Route("/auth", func(ar chi.Router) {
			ar.Get("/status", authHandler.Status)
			ar.Post("/setup", authHandler.Setup)
			ar.Post("/restore", backupHandler.RestoreInitial)
			ar.Post("/login", authHandler.Login)
			ar.Post("/logout", authHandler.Logout)
		})

		// Protected private routes
		api.Group(func(priv chi.Router) {
			priv.Use(authHandler.RequireAuth)

			priv.Get("/dashboard", dashboardHandler.Get)

			// Clients
			priv.Route("/clientes", func(cr chi.Router) {
				cr.Get("/", clienteHandler.List)
				cr.Post("/", clienteHandler.Create)
				cr.Post("/batch", clienteHandler.Batch)
				cr.Get("/{id}", clienteHandler.GetByID)
				cr.Put("/{id}", clienteHandler.Update)
				cr.Delete("/{id}", clienteHandler.Delete)
			})

			// Service Orders
			priv.Route("/ordens", func(or chi.Router) {
				or.Get("/", ordemHandler.List)
				or.Post("/", ordemHandler.Create)
				or.Post("/batch", ordemHandler.Batch)
				or.Get("/{id}", ordemHandler.GetByID)
				or.Get("/{id}/pdf", ordemHandler.DownloadPDF)
				or.Put("/{id}", ordemHandler.Update)
				or.Delete("/{id}", ordemHandler.Delete)
				or.Patch("/{id}/baixa", ordemHandler.BaixaPagamento)
			})

			// Quick Shortcuts (Read for all, mutations for Admin)
			priv.Route("/atalhos", func(atr chi.Router) {
				atr.Get("/", atalhoHandler.List)
				atr.With(authHandler.RequireAdmin).Post("/", atalhoHandler.Create)
				atr.With(authHandler.RequireAdmin).Put("/{id}", atalhoHandler.Update)
				atr.With(authHandler.RequireAdmin).Delete("/{id}", atalhoHandler.Delete)
			})

			// Admin-only operations: company configuration, user management, and backups
			priv.Group(func(admin chi.Router) {
				admin.Use(authHandler.RequireAdmin)

				admin.Put("/config", configHandler.Update)

				admin.Route("/usuarios", func(ur chi.Router) {
					ur.Get("/", authHandler.ListUsuarios)
					ur.Post("/", authHandler.CreateUsuario)
					ur.Put("/{id}", authHandler.UpdateUsuario)
					ur.Delete("/{id}", authHandler.DeleteUsuario)
				})

				admin.Route("/backups", func(br chi.Router) {
					br.Get("/", backupHandler.List)
					br.Post("/", backupHandler.Create)
					br.Put("/config", backupHandler.UpdateConfig)
					br.Post("/restore", backupHandler.Restore)
					br.Get("/{filename}/download", backupHandler.Download)
					br.Delete("/{filename}", backupHandler.Delete)
				})
			})
		})
	})

	// 5. Serve embedded Vue 3 SPA frontend (//go:embed)
	distFS := ui.DistFS()
	fileServer := http.FileServer(http.FS(distFS))

	r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet && req.Method != http.MethodHead {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		path := strings.TrimPrefix(req.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		// Check if file exists in embedded static filesystem
		f, err := distFS.Open(path)
		if err == nil {
			_ = f.Close()
			fileServer.ServeHTTP(w, req)
			return
		}

		// Fallback to index.html for SPA client-side routing
		indexFile, err := distFS.Open("index.html")
		if err != nil {
			http.NotFound(w, req)
			return
		}
		defer indexFile.Close()

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if stat, err := indexFile.Stat(); err == nil {
			http.ServeContent(w, req, "index.html", stat.ModTime(), indexFile.(io.ReadSeeker))
		} else {
			content, _ := io.ReadAll(indexFile)
			_, _ = w.Write(content)
		}
	}))

	// 6. Start HTTP Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8585"
	}
	addr := ":" + port

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("[OptiManager] Server listening at http://localhost%s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[OptiManager] HTTP server error: %v", err)
		}
	}()

	<-stop
	log.Println("[OptiManager] Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[OptiManager] Error during server shutdown: %v", err)
	}

	log.Println("[OptiManager] Server stopped successfully.")
}
