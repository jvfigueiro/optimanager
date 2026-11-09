package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"optimanager/internal/database"
	"optimanager/internal/models"
	"optimanager/internal/services/backup"
)

const maxBackupUploadBytes = 100 << 20 // 100 MB max upload size

var horarioRegex = regexp.MustCompile(`^([01]\d|2[0-3]):([0-5]\d)$`)

type BackupHandler struct {
	db      *database.DB
	service *backup.Service
}

func NewBackupHandler(db *database.DB, service *backup.Service) *BackupHandler {
	return &BackupHandler{
		db:      db,
		service: service,
	}
}

type BackupConfigPayload struct {
	BackupAutoEnabled          int    `json:"backup_auto_enabled"`
	BackupFrequencia           string `json:"backup_frequencia"`
	BackupDiaSemana            int    `json:"backup_dia_semana"`
	BackupHorario              string `json:"backup_horario"`
	BackupRetencaoAuto         int    `json:"backup_retencao_auto"`
	BackupRetencaoManual       int    `json:"backup_retencao_manual"`
	BackupRetencaoPrerestore   int    `json:"backup_retencao_prerestore"`
	BackupRetencaoPremigration int    `json:"backup_retencao_premigration"`
}

type BackupStatusResponse struct {
	Backups                    []backup.BackupInfo `json:"backups"`
	ExternalDir                string              `json:"external_dir"`
	BackupAutoEnabled          int                 `json:"backup_auto_enabled"`
	BackupFrequencia           string              `json:"backup_frequencia"`
	BackupDiaSemana            int                 `json:"backup_dia_semana"`
	BackupHorario              string              `json:"backup_horario"`
	BackupRetencaoAuto         int                 `json:"backup_retencao_auto"`
	BackupRetencaoManual       int                 `json:"backup_retencao_manual"`
	BackupRetencaoPrerestore   int                 `json:"backup_retencao_prerestore"`
	BackupRetencaoPremigration int                 `json:"backup_retencao_premigration"`
	BackupUltimoSucesso        string              `json:"backup_ultimo_sucesso"`
	BackupUltimoErro           string              `json:"backup_ultimo_erro"`
	BackupUltimoErroEm         string              `json:"backup_ultimo_erro_em"`
	BackupUltimoAvisoExterno   string              `json:"backup_ultimo_aviso_externo"`
	BackupUltimoAvisoExternoEm string              `json:"backup_ultimo_aviso_externo_em"`
}

// List returns all available local backups and current backup configuration/status.
func (h *BackupHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.ListBackups()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Falha ao listar arquivos de backup: "+err.Error())
		return
	}
	if list == nil {
		list = []backup.BackupInfo{}
	}

	cfg, err := h.db.GetConfig()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Falha ao carregar configurações de backup: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, BackupStatusResponse{
		Backups:                    list,
		ExternalDir:                h.service.ExternalDir(),
		BackupAutoEnabled:          cfg.BackupAutoEnabled,
		BackupFrequencia:           cfg.BackupFrequencia,
		BackupDiaSemana:            cfg.BackupDiaSemana,
		BackupHorario:              cfg.BackupHorario,
		BackupRetencaoAuto:         cfg.BackupRetencaoAuto,
		BackupRetencaoManual:       cfg.BackupRetencaoManual,
		BackupRetencaoPrerestore:   cfg.BackupRetencaoPrerestore,
		BackupRetencaoPremigration: cfg.BackupRetencaoPremigration,
		BackupUltimoSucesso:        cfg.BackupUltimoSucesso,
		BackupUltimoErro:           cfg.BackupUltimoErro,
		BackupUltimoErroEm:         cfg.BackupUltimoErroEm,
		BackupUltimoAvisoExterno:   cfg.BackupUltimoAvisoExterno,
		BackupUltimoAvisoExternoEm: cfg.BackupUltimoAvisoExternoEm,
	})
}

// Create triggers an immediate manual backup snapshot.
func (h *BackupHandler) Create(w http.ResponseWriter, r *http.Request) {
	info, err := h.service.CreateBackup(r.Context(), "manual")
	if err != nil {
		if errors.Is(err, backup.ErrInsufficientSpace) {
			respondError(w, http.StatusInsufficientStorage, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "Falha ao criar backup manual: "+err.Error())
		return
	}

	u := currentUser(r)
	ip := ExtractClientIP(r)
	if u != nil {
		_ = h.db.InsertAuditLog(u.ID, u.Nome, ip, "BACKUP_MANUAL", fmt.Sprintf("Backup manual criado: %s (%s)", info.Filename, info.TamanhoFmt))
	}

	respondJSON(w, http.StatusCreated, info)
}

// UpdateConfig updates the automated backup schedule and retention rules.
func (h *BackupHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var payload BackupConfigPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondError(w, http.StatusBadRequest, "Payload JSON inválido")
		return
	}

	freq := strings.ToLower(strings.TrimSpace(payload.BackupFrequencia))
	if freq != "diario" && freq != "semanal" {
		respondError(w, http.StatusBadRequest, "Frequência inválida: use 'diario' ou 'semanal'")
		return
	}

	if payload.BackupDiaSemana < 0 || payload.BackupDiaSemana > 6 {
		respondError(w, http.StatusBadRequest, "Dia da semana inválido (deve ser entre 0 e 6)")
		return
	}

	horario := strings.TrimSpace(payload.BackupHorario)
	if !horarioRegex.MatchString(horario) {
		respondError(w, http.StatusBadRequest, "Horário inválido: utilize o formato HH:MM (ex: 03:00)")
		return
	}

	if payload.BackupRetencaoAuto < 1 || payload.BackupRetencaoAuto > 60 {
		respondError(w, http.StatusBadRequest, "Retenção automática deve estar entre 1 e 60 arquivos")
		return
	}
	if payload.BackupRetencaoManual < 1 || payload.BackupRetencaoManual > 60 {
		respondError(w, http.StatusBadRequest, "Retenção manual deve estar entre 1 e 60 arquivos")
		return
	}
	if payload.BackupRetencaoPrerestore < 3 || payload.BackupRetencaoPrerestore > 60 {
		payload.BackupRetencaoPrerestore = 3
	}
	if payload.BackupRetencaoPremigration < 3 || payload.BackupRetencaoPremigration > 60 {
		payload.BackupRetencaoPremigration = 3
	}

	autoEnabled := 0
	if payload.BackupAutoEnabled != 0 {
		autoEnabled = 1
	}

	cfg, err := h.db.GetConfig()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	cfg.BackupAutoEnabled = autoEnabled
	cfg.BackupFrequencia = freq
	cfg.BackupDiaSemana = payload.BackupDiaSemana
	cfg.BackupHorario = horario
	cfg.BackupRetencaoAuto = payload.BackupRetencaoAuto
	cfg.BackupRetencaoManual = payload.BackupRetencaoManual
	cfg.BackupRetencaoPrerestore = payload.BackupRetencaoPrerestore
	cfg.BackupRetencaoPremigration = payload.BackupRetencaoPremigration

	if err := h.db.UpdateBackupConfig(cfg); err != nil {
		respondError(w, http.StatusInternalServerError, "Falha ao salvar configuração de backup: "+err.Error())
		return
	}

	u := currentUser(r)
	ip := ExtractClientIP(r)
	if u != nil {
		_ = h.db.InsertAuditLog(u.ID, u.Nome, ip, "BACKUP_CONFIG",
			fmt.Sprintf("Rotina de backup atualizada: ativo=%d, freq=%s, horario=%s, ret_auto=%d, ret_manual=%d",
				autoEnabled, freq, horario, payload.BackupRetencaoAuto, payload.BackupRetencaoManual))
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Configurações de backup salvas com sucesso"})
}

// Download streams a backup file to the browser as a binary attachment.
func (h *BackupHandler) Download(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	targetPath, err := h.service.GetBackupFilePath(filename)
	if err != nil {
		if errors.Is(err, backup.ErrBackupNotFound) {
			respondError(w, http.StatusNotFound, "Arquivo de backup não encontrado")
			return
		}
		respondError(w, http.StatusBadRequest, "Nome de arquivo de backup inválido")
		return
	}

	f, err := os.Open(targetPath)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Falha ao abrir arquivo de backup")
		return
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Falha ao ler metadados do arquivo de backup")
		return
	}

	u := currentUser(r)
	ip := ExtractClientIP(r)
	if u != nil {
		_ = h.db.InsertAuditLog(u.ID, u.Nome, ip, "BACKUP_DOWNLOAD", fmt.Sprintf("Download de backup: %s", fi.Name()))
	}

	w.Header().Set("Content-Type", "application/vnd.sqlite3")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fi.Name()))
	w.Header().Set("Cache-Control", "no-store")
	http.ServeContent(w, r, fi.Name(), fi.ModTime(), f)
}

// Delete removes a backup file from the server.
func (h *BackupHandler) Delete(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	if err := h.service.DeleteBackup(filename); err != nil {
		if errors.Is(err, backup.ErrBackupNotFound) {
			respondError(w, http.StatusNotFound, "Arquivo de backup não encontrado")
			return
		}
		respondError(w, http.StatusBadRequest, "Nome de arquivo inválido: "+err.Error())
		return
	}

	u := currentUser(r)
	ip := ExtractClientIP(r)
	if u != nil {
		_ = h.db.InsertAuditLog(u.ID, u.Nome, ip, "BACKUP_DELETE", fmt.Sprintf("Backup excluído: %s", filename))
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Backup removido com sucesso"})
}

// Restore restores the database either from an uploaded .db file (multipart/form-data)
// or from an existing backup filename on disk (application/json).
// Requires Admin authentication and automatically creates a pre-restore safety snapshot.
func (h *BackupHandler) Restore(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	ip := ExtractClientIP(r)
	var uid int64
	uname := "admin"
	if u != nil {
		uid = u.ID
		uname = u.Nome
	}

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		r.Body = http.MaxBytesReader(w, r.Body, maxBackupUploadBytes)
		if err := r.ParseMultipartForm(16 << 20); err != nil {
			respondError(w, http.StatusBadRequest, "Falha ao processar arquivo enviado (tamanho máximo: 100 MB)")
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			respondError(w, http.StatusBadRequest, "Nenhum arquivo de backup (.db) foi selecionado")
			return
		}
		defer file.Close()

		if err := h.service.RestoreFromReader(r.Context(), file, header.Filename, true, uid, uname, ip); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}

		respondJSON(w, http.StatusOK, map[string]string{
			"message": "Banco de dados restaurado com sucesso a partir do arquivo enviado!",
		})
		return
	}

	// JSON restore by existing filename
	var req struct {
		Filename string `json:"filename"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Filename) == "" {
		respondError(w, http.StatusBadRequest, "Informe o nome do arquivo de backup ou envie um arquivo .db")
		return
	}

	if err := h.service.RestoreFromFilename(r.Context(), strings.TrimSpace(req.Filename), true, uid, uname, ip); err != nil {
		if errors.Is(err, backup.ErrBackupNotFound) {
			respondError(w, http.StatusNotFound, "Arquivo de backup não encontrado")
			return
		}
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Banco de dados restaurado com sucesso!",
	})
}

// RestoreInitial allows uploading and restoring a .db backup file during the First-Run Welcome Wizard.
// Strictly restricted to uninitialized instances where CountUsuarios() == 0 (needs_setup == true).
func (h *BackupHandler) RestoreInitial(w http.ResponseWriter, r *http.Request) {
	count, err := h.db.CountUsuarios()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Erro ao verificar status de inicialização do banco")
		return
	}
	if count > 0 {
		respondError(w, http.StatusForbidden, "O sistema já possui usuários configurados. Faça login como Administrador para restaurar backups pela tela de Configurações.")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBackupUploadBytes)
	if err := r.ParseMultipartForm(16 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "Falha ao ler o arquivo enviado (limite máximo: 100 MB)")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Selecione um arquivo de backup (.db) válido para importar")
		return
	}
	defer file.Close()

	ip := ExtractClientIP(r)
	if err := h.service.RestoreFromReader(r.Context(), file, header.Filename, false, 0, "assistente_boas_vindas", ip); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message":   "Backup importado e restaurado com sucesso! Agora você pode entrar com suas credenciais.",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func currentUser(r *http.Request) *models.Usuario {
	if val := r.Context().Value(UserContextKey); val != nil {
		if u, ok := val.(*models.Usuario); ok {
			return u
		}
	}
	return nil
}
