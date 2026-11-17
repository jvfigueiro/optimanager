package database

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"optimanager/internal/models"
)

func normalizeRoleAndPerfil(role, perfil string) (string, string) {
	role = strings.TrimSpace(strings.ToLower(role))
	perfil = strings.TrimSpace(strings.ToLower(perfil))

	switch perfil {
	case "administrador":
		return "admin", "administrador"
	case "atendente":
		return "operador", "atendente"
	case "tecnico":
		return "operador", "tecnico"
	}

	switch role {
	case "operador", "atendente":
		return "operador", "atendente"
	case "tecnico":
		return "operador", "tecnico"
	default:
		return "admin", "administrador"
	}
}

// loadUserPermissions computes the effective permissions for a user by combining
// profile defaults (`perfil_permissoes`) with per-user overrides (`usuario_permissoes`).
func (d *DB) loadUserPermissions(u *models.Usuario) ([]string, error) {
	if u == nil || !u.Ativo {
		return []string{}, nil
	}

	permSet := make(map[string]bool)

	rows, err := d.Reader.Query(`SELECT permissao FROM perfil_permissoes WHERE perfil = ?`, u.Perfil)
	if err == nil {
		for rows.Next() {
			var p string
			if err := rows.Scan(&p); err == nil {
				permSet[p] = true
			}
		}
		rows.Close()
	}

	oRows, err := d.Reader.Query(`SELECT permissao, concedida FROM usuario_permissoes WHERE usuario_id = ?`, u.ID)
	if err == nil {
		for oRows.Next() {
			var p string
			var concedida int
			if err := oRows.Scan(&p, &concedida); err == nil {
				if concedida != 0 {
					permSet[p] = true
				} else {
					delete(permSet, p)
				}
			}
		}
		oRows.Close()
	}

	perms := make([]string, 0, len(permSet))
	for p, ok := range permSet {
		if ok {
			perms = append(perms, p)
		}
	}
	sort.Strings(perms)
	return perms, nil
}

// SetUsuarioPermissao sets or overrides a specific permission for a user.
func (d *DB) SetUsuarioPermissao(usuarioID int64, permissao string, concedida bool) error {
	val := 0
	if concedida {
		val = 1
	}
	_, err := d.Writer.Exec(
		`INSERT INTO usuario_permissoes (usuario_id, permissao, concedida) VALUES (?, ?, ?)
		 ON CONFLICT(usuario_id, permissao) DO UPDATE SET concedida = excluded.concedida`,
		usuarioID, strings.TrimSpace(permissao), val,
	)
	return err
}

// CountUsuarios returns the total count of active credentialed users in the database.
func (d *DB) CountUsuarios() (int, error) {
	var count int
	err := d.Reader.QueryRow(`SELECT COUNT(*) FROM usuarios WHERE ativo = 1 AND senha_hash != ''`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return count, nil
}

// ListUsuarios returns all registered users ordered by ID ascending.
func (d *DB) ListUsuarios() ([]models.Usuario, error) {
	rows, err := d.Reader.Query(
		`SELECT id, nome, COALESCE(email, ''), role, perfil, ativo, versao, created_at FROM usuarios ORDER BY id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	usuarios := make([]models.Usuario, 0, 8)
	for rows.Next() {
		var u models.Usuario
		var ativoInt int
		if err := rows.Scan(&u.ID, &u.Nome, &u.Email, &u.Role, &u.Perfil, &ativoInt, &u.Versao, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		u.Ativo = ativoInt != 0
		usuarios = append(usuarios, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range usuarios {
		perms, _ := d.loadUserPermissions(&usuarios[i])
		usuarios[i].Permissoes = perms
	}
	return usuarios, nil
}

// GetUsuarioByID fetches a single user by primary key ID along with effective permissions.
func (d *DB) GetUsuarioByID(id int64) (*models.Usuario, error) {
	var u models.Usuario
	var ativoInt int
	err := d.Reader.QueryRow(
		`SELECT id, nome, COALESCE(email, ''), senha_hash, role, perfil, ativo, versao, created_at FROM usuarios WHERE id = ?`,
		id,
	).Scan(&u.ID, &u.Nome, &u.Email, &u.SenhaHash, &u.Role, &u.Perfil, &ativoInt, &u.Versao, &u.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user #%d: %w", id, err)
	}
	u.Ativo = ativoInt != 0
	u.Permissoes, _ = d.loadUserPermissions(&u)
	return &u, nil
}

// GetUsuarioByEmail fetches a user by email address (case-insensitive).
func (d *DB) GetUsuarioByEmail(email string) (*models.Usuario, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return nil, nil
	}
	var u models.Usuario
	var ativoInt int
	err := d.Reader.QueryRow(
		`SELECT id, nome, COALESCE(email, ''), senha_hash, role, perfil, ativo, versao, created_at FROM usuarios WHERE email = ? COLLATE NOCASE`,
		email,
	).Scan(&u.ID, &u.Nome, &u.Email, &u.SenhaHash, &u.Role, &u.Perfil, &ativoInt, &u.Versao, &u.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	u.Ativo = ativoInt != 0
	u.Permissoes, _ = d.loadUserPermissions(&u)
	return &u, nil
}

// CreateUsuario hashes the password with bcrypt and stores a new active user record.
func (d *DB) CreateUsuario(nome, email, senha, role string) (*models.Usuario, error) {
	return d.CreateUsuarioWithProfile(nome, email, senha, role, "", true)
}

// CreateUsuarioWithProfile creates a user with explicit profile and active status.
// Inactive users (such as imported legacy technicians) may have an empty email and empty password.
func (d *DB) CreateUsuarioWithProfile(nome, email, senha, role, perfil string, ativo bool) (*models.Usuario, error) {
	nome = strings.TrimSpace(nome)
	email = strings.TrimSpace(strings.ToLower(email))
	normRole, normPerfil := normalizeRoleAndPerfil(role, perfil)

	if nome == "" {
		return nil, errors.New("user name is required")
	}

	var hashStr string
	if ativo {
		if email == "" || len(senha) < 4 {
			return nil, errors.New("name, email and password (minimum 4 characters) are required for active users")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to generate password hash: %w", err)
		}
		hashStr = string(hash)
	} else if strings.TrimSpace(senha) != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to generate password hash: %w", err)
		}
		hashStr = string(hash)
	}

	var emailParam interface{}
	if email != "" {
		emailParam = email
	}

	ativoInt := 0
	if ativo {
		ativoInt = 1
	}

	res, err := d.Writer.Exec(
		`INSERT INTO usuarios (nome, email, senha_hash, role, perfil, ativo, versao) VALUES (?, ?, ?, ?, ?, ?, 1)`,
		nome, emailParam, hashStr, normRole, normPerfil, ativoInt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return nil, errors.New("a user with this email already exists")
		}
		return nil, fmt.Errorf("failed to insert user: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return d.GetUsuarioByID(id)
}

// CreateInactiveLegacyTechnician creates an inactive technician record without login credentials.
func (d *DB) CreateInactiveLegacyTechnician(nome string) (*models.Usuario, error) {
	return d.CreateUsuarioWithProfile(nome, "", "", "operador", "tecnico", false)
}

// UpdateUsuario updates user profile details and optionally sets a new password.
func (d *DB) UpdateUsuario(id int64, nome, email, senha, role string) (*models.Usuario, error) {
	return d.UpdateUsuarioFull(id, &models.UpdateUsuarioRequest{
		Nome:  nome,
		Email: email,
		Senha: senha,
		Role:  role,
	})
}

// UpdateUsuarioFull updates user fields including optional profile, active status, and optimistic locking version.
func (d *DB) UpdateUsuarioFull(id int64, req *models.UpdateUsuarioRequest) (*models.Usuario, error) {
	existing, err := d.GetUsuarioByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("user not found")
	}

	if req.Versao > 0 && req.Versao != existing.Versao {
		return nil, ErrVersionConflict
	}

	nome := strings.TrimSpace(req.Nome)
	email := strings.TrimSpace(strings.ToLower(req.Email))

	roleInput := req.Role
	if roleInput == "" && req.Perfil == "" {
		roleInput = "operador"
	}
	normRole, normPerfil := normalizeRoleAndPerfil(roleInput, req.Perfil)

	ativo := existing.Ativo
	if req.Ativo != nil {
		ativo = *req.Ativo
	}

	if nome == "" || (ativo && email == "") {
		return nil, errors.New("name and email are required")
	}

	var emailParam interface{}
	if email != "" {
		emailParam = email
	}

	ativoInt := 0
	if ativo {
		ativoInt = 1
	}

	var res sql.Result
	if strings.TrimSpace(req.Senha) != "" {
		if len(req.Senha) < 4 {
			return nil, errors.New("new password must have at least 4 characters")
		}
		hash, errHash := bcrypt.GenerateFromPassword([]byte(req.Senha), bcrypt.DefaultCost)
		if errHash != nil {
			return nil, fmt.Errorf("failed to hash password: %w", errHash)
		}
		res, err = d.Writer.Exec(
			`UPDATE usuarios SET nome = ?, email = ?, senha_hash = ?, role = ?, perfil = ?, ativo = ?, versao = versao + 1 WHERE id = ? AND versao = ?`,
			nome, emailParam, string(hash), normRole, normPerfil, ativoInt, id, existing.Versao,
		)
	} else {
		res, err = d.Writer.Exec(
			`UPDATE usuarios SET nome = ?, email = ?, role = ?, perfil = ?, ativo = ?, versao = versao + 1 WHERE id = ? AND versao = ?`,
			nome, emailParam, normRole, normPerfil, ativoInt, id, existing.Versao,
		)
	}

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return nil, errors.New("another user with this email already exists")
		}
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		return nil, ErrVersionConflict
	}

	return d.GetUsuarioByID(id)
}

// DeleteUsuario deletes a user while strictly ensuring at least 1 active user remains in the system.
func (d *DB) DeleteUsuario(id int64) error {
	count, err := d.CountUsuarios()
	if err != nil {
		return err
	}
	if count <= 1 {
		return errors.New("cannot delete the only remaining user in the system")
	}

	res, err := d.Writer.Exec(`DELETE FROM usuarios WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("user not found")
	}
	return nil
}
