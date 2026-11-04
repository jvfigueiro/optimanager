package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"

	"optimanager/internal/database"
	"optimanager/internal/models"
)

type contextKey string

const UserContextKey contextKey = "auth_user"
const CookieName = "optimanager_token"

type tokenPayload struct {
	UserID int64  `json:"uid"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Exp    int64  `json:"exp"`
}

type AuthHandler struct {
	db     *database.DB
	secret []byte
}

func NewAuthHandler(db *database.DB) *AuthHandler {
	sec := os.Getenv("JWT_SECRET")
	if sec == "" {
		sec = "optimanager-default-secret-key-2026-hmac-sha256"
	}
	return &AuthHandler{
		db:     db,
		secret: []byte(sec),
	}
}

func (h *AuthHandler) generateToken(u *models.Usuario) (string, error) {
	payload := tokenPayload{
		UserID: u.ID,
		Email:  u.Email,
		Role:   u.Role,
		Exp:    time.Now().Add(7 * 24 * time.Hour).Unix(), // 7 dias de validade
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(b)
	mac := hmac.New(sha256.New, h.secret)
	mac.Write([]byte(encodedPayload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return encodedPayload + "." + sig, nil
}

func (h *AuthHandler) verifyToken(tokenStr string) (*tokenPayload, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 2 {
		return nil, errors.New("formato de token inválido")
	}

	mac := hmac.New(sha256.New, h.secret)
	mac.Write([]byte(parts[0]))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(parts[1]), []byte(expectedSig)) {
		return nil, errors.New("assinatura de token inválida")
	}

	rawPayload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, errors.New("erro ao decodificar payload do token")
	}

	var payload tokenPayload
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return nil, errors.New("payload do token malformado")
	}

	if time.Now().Unix() > payload.Exp {
		return nil, errors.New("token expirado")
	}

	return &payload, nil
}

func (h *AuthHandler) extractUserFromRequest(r *http.Request) *models.Usuario {
	var tokenStr string

	// 1. Tenta header Authorization: Bearer <token>
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		tokenStr = strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	}

	// 2. Tenta Cookie HttpOnly
	if tokenStr == "" {
		if cookie, err := r.Cookie(CookieName); err == nil {
			tokenStr = cookie.Value
		}
	}

	// 3. Tenta query parameter ?token= (para links de download PDF diretos se necessário)
	if tokenStr == "" {
		tokenStr = r.URL.Query().Get("token")
	}

	if tokenStr == "" {
		return nil
	}

	payload, err := h.verifyToken(tokenStr)
	if err != nil {
		return nil
	}

	u, err := h.db.GetUsuarioByID(payload.UserID)
	if err != nil || u == nil {
		return nil
	}

	return u
}

func (h *AuthHandler) setTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   7 * 24 * 3600,
	})
}

// Middleware RequireAuth protege rotas privadas da API.
func (h *AuthHandler) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := h.extractUserFromRequest(r)
		if u == nil {
			respondError(w, http.StatusUnauthorized, "Não autenticado. Faça login para continuar.")
			return
		}
		ctx := context.WithValue(r.Context(), UserContextKey, u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Status retorna se o sistema precisa de setup inicial (0 usuários) e se o usuário atual está autenticado.
func (h *AuthHandler) Status(w http.ResponseWriter, r *http.Request) {
	count, err := h.db.CountUsuarios()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if count == 0 {
		respondJSON(w, http.StatusOK, models.AuthStatusResponse{
			NeedsSetup:    true,
			Authenticated: false,
		})
		return
	}

	u := h.extractUserFromRequest(r)
	if u != nil {
		respondJSON(w, http.StatusOK, models.AuthStatusResponse{
			NeedsSetup:    false,
			Authenticated: true,
			Usuario:       u,
		})
		return
	}

	respondJSON(w, http.StatusOK, models.AuthStatusResponse{
		NeedsSetup:    false,
		Authenticated: false,
	})
}

// Setup realiza o cadastro do administrador no primeiro acesso (quando há 0 usuários).
func (h *AuthHandler) Setup(w http.ResponseWriter, r *http.Request) {
	count, err := h.db.CountUsuarios()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if count > 0 {
		respondError(w, http.StatusForbidden, "O administrador inicial já foi configurado. Faça login normalmente.")
		return
	}

	var req models.SetupAdminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Payload JSON inválido")
		return
	}

	if strings.TrimSpace(req.Nome) == "" || strings.TrimSpace(req.Email) == "" || len(req.Senha) < 4 {
		respondError(w, http.StatusBadRequest, "Preencha nome, e-mail válido e uma senha de no mínimo 4 caracteres")
		return
	}

	u, err := h.db.CreateUsuario(req.Nome, req.Email, req.Senha, "admin")
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Se dados da empresa foram informados no assistente de primeiro uso, atualiza a tabela config
	if cfg, err := h.db.GetConfig(); err == nil && cfg != nil {
		cfg.Responsavel = strings.TrimSpace(req.Nome)
		if strings.TrimSpace(req.Fantasia) != "" {
			cfg.Fantasia = strings.TrimSpace(req.Fantasia)
		}
		if strings.TrimSpace(req.RazaoSocial) != "" {
			cfg.RazaoSocial = strings.TrimSpace(req.RazaoSocial)
		}
		if strings.TrimSpace(req.CNPJ) != "" {
			cfg.CNPJ = strings.TrimSpace(req.CNPJ)
		}
		if strings.TrimSpace(req.Fones) != "" {
			cfg.Fones = strings.TrimSpace(req.Fones)
		}
		if strings.TrimSpace(req.EmailContato) != "" {
			cfg.EmailContato = strings.TrimSpace(req.EmailContato)
		}
		if strings.TrimSpace(req.Endereco) != "" {
			cfg.Endereco = strings.TrimSpace(req.Endereco)
		}
		if strings.TrimSpace(req.Cidade) != "" {
			cfg.Cidade = strings.TrimSpace(req.Cidade)
		}
		_ = h.db.UpdateConfig(cfg)
	}

	token, err := h.generateToken(u)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Erro ao gerar sessão")
		return
	}

	h.setTokenCookie(w, token)
	respondJSON(w, http.StatusCreated, models.AuthResponse{
		Token:   token,
		Usuario: u,
	})
}

// Login autentica um usuário existente com e-mail e senha.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Payload JSON inválido")
		return
	}

	u, err := h.db.GetUsuarioByEmail(req.Email)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Erro ao verificar credenciais")
		return
	}
	if u == nil {
		respondError(w, http.StatusUnauthorized, "E-mail ou senha incorretos")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.SenhaHash), []byte(req.Senha)); err != nil {
		respondError(w, http.StatusUnauthorized, "E-mail ou senha incorretos")
		return
	}

	token, err := h.generateToken(u)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Erro ao gerar token de acesso")
		return
	}

	h.setTokenCookie(w, token)
	respondJSON(w, http.StatusOK, models.AuthResponse{
		Token:   token,
		Usuario: u,
	})
}

// Logout encerra a sessão limpando o cookie de autenticação.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	respondJSON(w, http.StatusOK, map[string]string{"message": "Sessão encerrada com sucesso"})
}

// ListUsuarios retorna a lista de todos os usuários para a tela de configurações.
func (h *AuthHandler) ListUsuarios(w http.ResponseWriter, r *http.Request) {
	usuarios, err := h.db.ListUsuarios()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, usuarios)
}

// CreateUsuario adiciona um novo usuário a partir da tela de configurações.
func (h *AuthHandler) CreateUsuario(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUsuarioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Payload JSON inválido")
		return
	}

	u, err := h.db.CreateUsuario(req.Nome, req.Email, req.Senha, req.Role)
	if err != nil {
		if strings.Contains(err.Error(), "já existe") {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, u)
}

// UpdateUsuario edita os dados (e opcionalmente a senha) de um usuário existente.
func (h *AuthHandler) UpdateUsuario(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "ID de usuário inválido")
		return
	}

	var req models.UpdateUsuarioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Payload JSON inválido")
		return
	}

	u, err := h.db.UpdateUsuario(id, req.Nome, req.Email, req.Senha, req.Role)
	if err != nil {
		if strings.Contains(err.Error(), "não encontrado") {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		if strings.Contains(err.Error(), "já existe") {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, u)
}

// DeleteUsuario remove um usuário do sistema.
func (h *AuthHandler) DeleteUsuario(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "ID de usuário inválido")
		return
	}

	// Impede que o usuário logado exclua a própria conta ativa
	if currentUser, ok := r.Context().Value(UserContextKey).(*models.Usuario); ok && currentUser != nil {
		if currentUser.ID == id {
			respondError(w, http.StatusBadRequest, "Você não pode excluir sua própria conta em uso.")
			return
		}
	}

	if err := h.db.DeleteUsuario(id); err != nil {
		if strings.Contains(err.Error(), "não encontrado") {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Usuário removido com sucesso"})
}

