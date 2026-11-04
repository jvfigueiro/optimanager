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

// CountUsuarios retorna o total de usuários cadastrados no sistema.
func (d *DB) CountUsuarios() (int, error) {
	var count int
	err := d.Reader.QueryRow(`SELECT COUNT(*) FROM usuarios`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("erro ao contar usuários: %w", err)
	}
	return count, nil
}

// ListUsuarios retorna todos os usuários ordenados por ID.
func (d *DB) ListUsuarios() ([]models.Usuario, error) {
	rows, err := d.Reader.Query(
		`SELECT id, nome, email, role, created_at FROM usuarios ORDER BY id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar usuários: %w", err)
	}
	defer rows.Close()

	usuarios := make([]models.Usuario, 0, 8)
	for rows.Next() {
		var u models.Usuario
		if err := rows.Scan(&u.ID, &u.Nome, &u.Email, &u.Role, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("erro ao ler usuário: %w", err)
		}
		usuarios = append(usuarios, u)
	}
	return usuarios, rows.Err()
}

// GetUsuarioByID busca um usuário pelo ID.
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
		return nil, fmt.Errorf("erro ao buscar usuário #%d: %w", id, err)
	}
	return &u, nil
}

// GetUsuarioByEmail busca um usuário pelo e-mail (case-insensitive).
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
		return nil, fmt.Errorf("erro ao buscar usuário por e-mail: %w", err)
	}
	return &u, nil
}

// CreateUsuario cria um novo usuário com senha criptografada via Bcrypt.
func (d *DB) CreateUsuario(nome, email, senha, role string) (*models.Usuario, error) {
	nome = strings.TrimSpace(nome)
	email = strings.TrimSpace(strings.ToLower(email))
	role = strings.TrimSpace(strings.ToLower(role))

	if nome == "" || email == "" || len(senha) < 4 {
		return nil, errors.New("nome, e-mail e senha (mínimo 4 caracteres) são obrigatórios")
	}
	if role != "admin" && role != "operador" {
		role = "admin"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar hash de senha: %w", err)
	}

	res, err := d.Writer.Exec(
		`INSERT INTO usuarios (nome, email, senha_hash, role) VALUES (?, ?, ?, ?)`,
		nome, email, string(hash), role,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return nil, errors.New("já existe um usuário cadastrado com este e-mail")
		}
		return nil, fmt.Errorf("erro ao cadastrar usuário: %w", err)
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

// UpdateUsuario atualiza nome, e-mail, perfil e opcionalmente a senha de um usuário.
func (d *DB) UpdateUsuario(id int64, nome, email, senha, role string) (*models.Usuario, error) {
	nome = strings.TrimSpace(nome)
	email = strings.TrimSpace(strings.ToLower(email))
	role = strings.TrimSpace(strings.ToLower(role))

	if nome == "" || email == "" {
		return nil, errors.New("nome e e-mail são obrigatórios")
	}
	if role != "admin" && role != "operador" {
		role = "operador"
	}

	var res sql.Result
	var err error

	if strings.TrimSpace(senha) != "" {
		if len(senha) < 4 {
			return nil, errors.New("a nova senha deve ter pelo menos 4 caracteres")
		}
		hash, errHash := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
		if errHash != nil {
			return nil, fmt.Errorf("erro ao criptografar senha: %w", errHash)
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
			return nil, errors.New("já existe outro usuário com este e-mail")
		}
		return nil, fmt.Errorf("erro ao atualizar usuário: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		return nil, errors.New("usuário não encontrado")
	}

	return d.GetUsuarioByID(id)
}

// DeleteUsuario remove um usuário do sistema garantindo que sempre reste ao menos 1 usuário ativo.
func (d *DB) DeleteUsuario(id int64) error {
	count, err := d.CountUsuarios()
	if err != nil {
		return err
	}
	if count <= 1 {
		return errors.New("não é permitido remover o único usuário do sistema")
	}

	res, err := d.Writer.Exec(`DELETE FROM usuarios WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("erro ao remover usuário: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("usuário não encontrado")
	}
	return nil
}

