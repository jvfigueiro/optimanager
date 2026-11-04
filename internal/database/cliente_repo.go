package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"optimanager/internal/models"
)

// ListClientes queries clients with optional search by name, document, or contact.
func (d *DB) ListClientes(search string) ([]models.Cliente, error) {
	search = strings.TrimSpace(search)
	var rows *sql.Rows
	var err error

	if search == "" {
		rows, err = d.Reader.Query(
			`SELECT id, nome, doc, contato, desde_ano, created_at 
			 FROM clientes 
			 ORDER BY nome COLLATE NOCASE ASC`,
		)
	} else {
		pattern := "%" + search + "%"
		rows, err = d.Reader.Query(
			`SELECT id, nome, doc, contato, desde_ano, created_at 
			 FROM clientes 
			 WHERE nome LIKE ? OR doc LIKE ? OR contato LIKE ?
			 ORDER BY nome COLLATE NOCASE ASC`,
			pattern, pattern, pattern,
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query clients: %w", err)
	}
	defer rows.Close()

	clientes := make([]models.Cliente, 0, 16)
	for rows.Next() {
		var c models.Cliente
		if err := rows.Scan(&c.ID, &c.Nome, &c.Doc, &c.Contato, &c.DesdeAno, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan client: %w", err)
		}
		clientes = append(clientes, c)
	}

	return clientes, rows.Err()
}

// GetClienteByID queries a single client by ID.
func (d *DB) GetClienteByID(id int64) (*models.Cliente, error) {
	var c models.Cliente
	err := d.Reader.QueryRow(
		`SELECT id, nome, doc, contato, desde_ano, created_at 
		 FROM clientes 
		 WHERE id = ?`,
		id,
	).Scan(&c.ID, &c.Nome, &c.Doc, &c.Contato, &c.DesdeAno, &c.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query client #%d: %w", id, err)
	}
	return &c, nil
}

// GetOrCreateCliente finds an existing client by name or inserts a new one within a transaction or write connection.
func (d *DB) GetOrCreateCliente(tx *sql.Tx, nome, doc, contato string) (*models.Cliente, error) {
	nome = strings.TrimSpace(nome)
	if nome == "" {
		return nil, errors.New("client name cannot be empty")
	}
	doc = strings.TrimSpace(doc)
	contato = strings.TrimSpace(contato)

	var c models.Cliente
	var err error

	queryExisting := `SELECT id, nome, doc, contato, desde_ano, created_at 
	                  FROM clientes 
	                  WHERE nome = ? COLLATE NOCASE`

	if tx != nil {
		err = tx.QueryRow(queryExisting, nome).Scan(&c.ID, &c.Nome, &c.Doc, &c.Contato, &c.DesdeAno, &c.CreatedAt)
	} else {
		err = d.Writer.QueryRow(queryExisting, nome).Scan(&c.ID, &c.Nome, &c.Doc, &c.Contato, &c.DesdeAno, &c.CreatedAt)
	}

	if err == nil {
		// Client exists; update empty fields if new details were supplied
		needsUpdate := false
		if c.Doc == "" && doc != "" {
			c.Doc = doc
			needsUpdate = true
		}
		if c.Contato == "" && contato != "" {
			c.Contato = contato
			needsUpdate = true
		}

		if needsUpdate {
			updateQuery := `UPDATE clientes SET doc = ?, contato = ? WHERE id = ?`
			if tx != nil {
				_, _ = tx.Exec(updateQuery, c.Doc, c.Contato, c.ID)
			} else {
				_, _ = d.Writer.Exec(updateQuery, c.Doc, c.Contato, c.ID)
			}
		}

		return &c, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to check existing client: %w", err)
	}

	// Create new client with current year
	desdeAno := time.Now().Year()
	insertQuery := `INSERT INTO clientes (nome, doc, contato, desde_ano) VALUES (?, ?, ?, ?)`

	var res sql.Result
	if tx != nil {
		res, err = tx.Exec(insertQuery, nome, doc, contato, desdeAno)
	} else {
		res, err = d.Writer.Exec(insertQuery, nome, doc, contato, desdeAno)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to auto-create client: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created client id: %w", err)
	}

	return &models.Cliente{
		ID:        id,
		Nome:      nome,
		Doc:       doc,
		Contato:   contato,
		DesdeAno:  desdeAno,
		CreatedAt: time.Now(),
	}, nil
}

// CreateCliente manually creates a new client.
func (d *DB) CreateCliente(c *models.Cliente) error {
	c.Nome = strings.TrimSpace(c.Nome)
	if c.Nome == "" {
		return errors.New("client name is required")
	}
	c.Doc = strings.TrimSpace(c.Doc)
	c.Contato = strings.TrimSpace(c.Contato)
	if c.DesdeAno <= 0 {
		c.DesdeAno = time.Now().Year()
	}

	res, err := d.Writer.Exec(
		`INSERT INTO clientes (nome, doc, contato, desde_ano) VALUES (?, ?, ?, ?)`,
		c.Nome, c.Doc, c.Contato, c.DesdeAno,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return errors.New("a client with this name already exists")
		}
		return fmt.Errorf("failed to create client: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	c.ID = id
	c.CreatedAt = time.Now()
	return nil
}

// UpdateCliente updates client information.
func (d *DB) UpdateCliente(c *models.Cliente) error {
	c.Nome = strings.TrimSpace(c.Nome)
	if c.Nome == "" {
		return errors.New("client name is required")
	}
	c.Doc = strings.TrimSpace(c.Doc)
	c.Contato = strings.TrimSpace(c.Contato)

	res, err := d.Writer.Exec(
		`UPDATE clientes 
		 SET nome = ?, doc = ?, contato = ?, desde_ano = ? 
		 WHERE id = ?`,
		c.Nome, c.Doc, c.Contato, c.DesdeAno, c.ID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return errors.New("another client with this name already exists")
		}
		return fmt.Errorf("failed to update client: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("client not found")
	}
	return nil
}

// DeleteCliente removes a client, failing if foreign key constraints exist on orders.
func (d *DB) DeleteCliente(id int64) error {
	res, err := d.Writer.Exec(`DELETE FROM clientes WHERE id = ?`, id)
	if err != nil {
		if strings.Contains(err.Error(), "FOREIGN KEY") || strings.Contains(err.Error(), "constraint failed") {
			return errors.New("cannot delete client with linked service orders")
		}
		return fmt.Errorf("failed to delete client: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("client not found")
	}
	return nil
}
