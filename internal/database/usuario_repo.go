package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"optimanager/internal/models"
)

// CountUsuarios returns the total count of registered users in the database.
func (d *DB) CountUsuarios() (int, error) {
	var count int
	err := d.Reader.QueryRow(`SELECT COUNT(*) FROM usuarios`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return count, nil
}

// ListUsuarios returns all registered users ordered by ID ascending.
func (d *DB) ListUsuarios() ([]models.Usuario, error) {
	rows, err := d.Reader.Query(
		`SELECT id, nome, email, role, created_at FROM usuarios ORDER BY id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	usuarios := make([]models.Usuario, 0, 8)
	for rows.Next() {
		var u models.Usuario
		if err := rows.Scan(&u.ID, &u.Nome, &u.Email, &u.Role, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		usuarios = append(usuarios, u)
	}
	return usuarios, rows.Err()
}

// GetUsuarioByID fetches a single user by primary key ID.
func (d *DB) GetUsuarioByID(id int64) (*models.Usuario, error) {
	var u models.Usuario
	err := d.Reader.QueryRow(
		`SELECT id, nome, email, senha_hash, role, created_at FROM usuarios WHERE id = ?`,
		id,
	).Scan(&u.ID, &u.Nome, &u.Email, &u.SenhaHash, &u.Role, &u.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user #%d: %w", id, err)
	}
	return &u, nil
}

// GetUsuarioByEmail fetches a user by email address (case-insensitive).
func (d *DB) GetUsuarioByEmail(email string) (*models.Usuario, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	var u models.Usuario
	err := d.Reader.QueryRow(
		`SELECT id, nome, email, senha_hash, role, created_at FROM usuarios WHERE email = ? COLLATE NOCASE`,
		email,
	).Scan(&u.ID, &u.Nome, &u.Email, &u.SenhaHash, &u.Role, &u.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &u, nil
}

// CreateUsuario hashes the password with bcrypt and stores a new user record.
func (d *DB) CreateUsuario(nome, email, senha, role string) (*models.Usuario, error) {
	nome = strings.TrimSpace(nome)
	email = strings.TrimSpace(strings.ToLower(email))
	role = strings.TrimSpace(strings.ToLower(role))

	if nome == "" || email == "" || len(senha) < 4 {
		return nil, errors.New("name, email and password (minimum 4 characters) are required")
	}
	if role != "admin" && role != "operador" {
		role = "admin"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to generate password hash: %w", err)
	}

	res, err := d.Writer.Exec(
		`INSERT INTO usuarios (nome, email, senha_hash, role) VALUES (?, ?, ?, ?)`,
		nome, email, string(hash), role,
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

	return &models.Usuario{
		ID:        id,
		Nome:      nome,
		Email:     email,
		Role:      role,
		CreatedAt: time.Now(),
	}, nil
}

// UpdateUsuario updates user profile details and optionally sets a new password.
func (d *DB) UpdateUsuario(id int64, nome, email, senha, role string) (*models.Usuario, error) {
	nome = strings.TrimSpace(nome)
	email = strings.TrimSpace(strings.ToLower(email))
	role = strings.TrimSpace(strings.ToLower(role))

	if nome == "" || email == "" {
		return nil, errors.New("name and email are required")
	}
	if role != "admin" && role != "operador" {
		role = "operador"
	}

	var res sql.Result
	var err error

	if strings.TrimSpace(senha) != "" {
		if len(senha) < 4 {
			return nil, errors.New("new password must have at least 4 characters")
		}
		hash, errHash := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
		if errHash != nil {
			return nil, fmt.Errorf("failed to hash password: %w", errHash)
		}
		res, err = d.Writer.Exec(
			`UPDATE usuarios SET nome = ?, email = ?, senha_hash = ?, role = ? WHERE id = ?`,
			nome, email, string(hash), role, id,
		)
	} else {
		res, err = d.Writer.Exec(
			`UPDATE usuarios SET nome = ?, email = ?, role = ? WHERE id = ?`,
			nome, email, role, id,
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
		return nil, errors.New("user not found")
	}

	return d.GetUsuarioByID(id)
}

// DeleteUsuario deletes a user while strictly ensuring at least 1 user remains in the system.
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
