package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"optimanager/internal/models"
)

// ListOrdens queries service orders with client join, filtering and textual search.
func (d *DB) ListOrdens(filtroStatus, busca string) ([]models.OrdemServico, error) {
	filtroStatus = strings.TrimSpace(strings.ToUpper(filtroStatus))
	busca = strings.TrimSpace(busca)

	query := `SELECT os.id, os.data_emissao, os.status, os.cliente_id,
	                 c.nome, c.doc, c.contato,
	                 os.equipamento, os.servico, os.obs_tecnicas, os.valor_total,
	                 os.pagamento_status, os.pagamento_forma, os.pago_em, os.created_at
	          FROM ordens_servico os
	          INNER JOIN clientes c ON c.id = os.cliente_id
	          WHERE 1=1`

	args := make([]interface{}, 0, 4)

	switch filtroStatus {
	case "", "TODOS":
		// No filter
	case "ABERTOS":
		query += ` AND os.status != 'Concluído'`
	case "CONCLUIDOS":
		query += ` AND os.status = 'Concluído'`
	case "PENDENTE":
		query += ` AND os.status = 'Concluído' AND os.pagamento_status = 'Pendente'`
	default:
		query += ` AND os.status = ?`
		args = append(args, filtroStatus)
	}

	if busca != "" {
		pattern := "%" + busca + "%"
		query += ` AND (c.nome LIKE ? OR os.equipamento LIKE ? OR os.servico LIKE ? OR os.obs_tecnicas LIKE ?)`
		args = append(args, pattern, pattern, pattern, pattern)
	}

	query += ` ORDER BY os.id DESC`

	rows, err := d.Reader.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query service orders: %w", err)
	}
	defer rows.Close()

	var cnpj, fantasia string
	_ = d.Reader.QueryRow(`SELECT cnpj, fantasia FROM config WHERE id = 1`).Scan(&cnpj, &fantasia)

	ordens := make([]models.OrdemServico, 0, 32)
	for rows.Next() {
		var os models.OrdemServico
		var pagoEm sql.NullTime

		err := rows.Scan(
			&os.ID, &os.DataEmissao, &os.Status, &os.ClienteID,
			&os.ClienteNome, &os.ClienteDoc, &os.ClienteContato,
			&os.Equipamento, &os.Servico, &os.ObsTecnicas, &os.ValorTotal,
			&os.PagamentoStatus, &os.PagamentoForma, &pagoEm, &os.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan service order: %w", err)
		}

		if pagoEm.Valid {
			os.PagoEm = &pagoEm.Time
		}

		os.WhatsAppURL = os.BuildWhatsAppURL(cnpj, fantasia)
		ordens = append(ordens, os)
	}

	return ordens, rows.Err()
}

// GetOrdemByID queries a single order with joined client data.
func (d *DB) GetOrdemByID(id int64) (*models.OrdemServico, error) {
	query := `SELECT os.id, os.data_emissao, os.status, os.cliente_id,
	                 c.nome, c.doc, c.contato,
	                 os.equipamento, os.servico, os.obs_tecnicas, os.valor_total,
	                 os.pagamento_status, os.pagamento_forma, os.pago_em, os.created_at
	          FROM ordens_servico os
	          INNER JOIN clientes c ON c.id = os.cliente_id
	          WHERE os.id = ?`

	var os models.OrdemServico
	var pagoEm sql.NullTime

	err := d.Reader.QueryRow(query, id).Scan(
		&os.ID, &os.DataEmissao, &os.Status, &os.ClienteID,
		&os.ClienteNome, &os.ClienteDoc, &os.ClienteContato,
		&os.Equipamento, &os.Servico, &os.ObsTecnicas, &os.ValorTotal,
		&os.PagamentoStatus, &os.PagamentoForma, &pagoEm, &os.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get order #%d: %w", id, err)
	}

	if pagoEm.Valid {
		os.PagoEm = &pagoEm.Time
	}

	var cnpj, fantasia string
	_ = d.Reader.QueryRow(`SELECT cnpj, fantasia FROM config WHERE id = 1`).Scan(&cnpj, &fantasia)
	os.WhatsAppURL = os.BuildWhatsAppURL(cnpj, fantasia)

	return &os, nil
}

// CreateOrdem creates a new service order within a transaction with auto-client creation and VT shortcut support.
func (d *DB) CreateOrdem(req *models.CreateOrdemRequest) (*models.OrdemServico, error) {
	if strings.TrimSpace(req.Servico) == "" {
		return nil, errors.New("service description is required")
	}

	tx, err := d.Writer.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Resolve or auto-create client
	var cliente *models.Cliente
	if req.ClienteID > 0 {
		cliente = &models.Cliente{ID: req.ClienteID}
		err = tx.QueryRow(`SELECT nome, doc, contato, desde_ano, created_at FROM clientes WHERE id = ?`, req.ClienteID).
			Scan(&cliente.Nome, &cliente.Doc, &cliente.Contato, &cliente.DesdeAno, &cliente.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("client #%d not found: %w", req.ClienteID, err)
		}
	} else {
		if strings.TrimSpace(req.ClienteNome) == "" {
			return nil, errors.New("client name or ID is required")
		}
		cliente, err = d.GetOrCreateCliente(tx, req.ClienteNome, req.ClienteDoc, req.ClienteContato)
		if err != nil {
			return nil, err
		}
	}

	// 2. Business rules: Technical Visit (VT) shortcut
	equipamento := strings.TrimSpace(req.Equipamento)
	servico := strings.TrimSpace(req.Servico)
	valor := req.ValorTotal

	if req.IsVT {
		if equipamento == "" || strings.EqualFold(equipamento, "N/A") {
			equipamento = "N/A"
		}
		if !strings.HasPrefix(strings.ToUpper(servico), "VT - ") {
			servico = "VT - " + servico
		}
		if valor <= 0 {
			var vtPadrao float64
			_ = tx.QueryRow(`SELECT vt_valor_padrao FROM config WHERE id = 1`).Scan(&vtPadrao)
			if vtPadrao > 0 {
				valor = vtPadrao
			} else {
				valor = 80.00
			}
		}
	} else if equipamento == "" {
		equipamento = "N/A"
	}

	// 3. Field defaults
	dataEmissao := strings.TrimSpace(req.DataEmissao)
	if dataEmissao == "" {
		dataEmissao = time.Now().Format("2006-01-02")
	}

	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "Avaliação"
	}

	pagamentoStatus := strings.TrimSpace(req.PagamentoStatus)
	if pagamentoStatus == "" {
		pagamentoStatus = "Pendente"
	}

	pagamentoForma := strings.TrimSpace(req.PagamentoForma)
	if pagamentoForma == "" {
		pagamentoForma = "Pix"
	}

	var pagoEm sql.NullTime
	if pagamentoStatus == "Pago" {
		pagoEm = sql.NullTime{Time: time.Now(), Valid: true}
	}

	// 4. Insert into database
	query := `INSERT INTO ordens_servico (
		data_emissao, status, cliente_id, equipamento, servico, 
		obs_tecnicas, valor_total, pagamento_status, pagamento_forma, pago_em
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := tx.Exec(query,
		dataEmissao,
		status,
		cliente.ID,
		equipamento,
		servico,
		strings.TrimSpace(req.ObsTecnicas),
		valor,
		pagamentoStatus,
		pagamentoForma,
		pagoEm,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert service order: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get generated order id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	var cnpj, fantasia string
	_ = d.Reader.QueryRow(`SELECT cnpj, fantasia FROM config WHERE id = 1`).Scan(&cnpj, &fantasia)

	novaOS := &models.OrdemServico{
		ID:              id,
		DataEmissao:     dataEmissao,
		Status:          status,
		ClienteID:       cliente.ID,
		ClienteNome:     cliente.Nome,
		ClienteDoc:      cliente.Doc,
		ClienteContato:  cliente.Contato,
		Equipamento:     equipamento,
		Servico:         servico,
		ObsTecnicas:     req.ObsTecnicas,
		ValorTotal:      valor,
		PagamentoStatus: pagamentoStatus,
		PagamentoForma:  pagamentoForma,
		CreatedAt:       time.Now(),
	}
	if pagoEm.Valid {
		novaOS.PagoEm = &pagoEm.Time
	}
	novaOS.WhatsAppURL = novaOS.BuildWhatsAppURL(cnpj, fantasia)

	return novaOS, nil
}

// UpdateOrdem updates an existing service order record.
func (d *DB) UpdateOrdem(id int64, req *models.UpdateOrdemRequest) (*models.OrdemServico, error) {
	if strings.TrimSpace(req.Servico) == "" {
		return nil, errors.New("service description is required")
	}

	tx, err := d.Writer.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var osExistente models.OrdemServico
	var pagoEmExistente sql.NullTime
	err = tx.QueryRow(`SELECT id, status, cliente_id, pagamento_status, pago_em FROM ordens_servico WHERE id = ?`, id).
		Scan(&osExistente.ID, &osExistente.Status, &osExistente.ClienteID, &osExistente.PagamentoStatus, &pagoEmExistente)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("service order not found")
		}
		return nil, err
	}

	clienteID := osExistente.ClienteID
	if req.ClienteID > 0 {
		clienteID = req.ClienteID
	} else if strings.TrimSpace(req.ClienteNome) != "" {
		cli, err := d.GetOrCreateCliente(tx, req.ClienteNome, req.ClienteDoc, req.ClienteContato)
		if err != nil {
			return nil, err
		}
		clienteID = cli.ID
	}

	equipamento := strings.TrimSpace(req.Equipamento)
	if equipamento == "" {
		equipamento = "N/A"
	}

	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = osExistente.Status
	}

	pagamentoStatus := strings.TrimSpace(req.PagamentoStatus)
	if pagamentoStatus == "" {
		pagamentoStatus = osExistente.PagamentoStatus
	}

	pagamentoForma := strings.TrimSpace(req.PagamentoForma)
	if pagamentoForma == "" {
		pagamentoForma = "Pix"
	}

	pagoEm := pagoEmExistente
	if pagamentoStatus == "Pago" && !pagoEm.Valid {
		pagoEm = sql.NullTime{Time: time.Now(), Valid: true}
	} else if pagamentoStatus == "Pendente" {
		pagoEm = sql.NullTime{Valid: false}
	}

	dataEmissao := strings.TrimSpace(req.DataEmissao)
	if dataEmissao == "" {
		dataEmissao = time.Now().Format("2006-01-02")
	}

	updateQuery := `UPDATE ordens_servico SET 
		data_emissao = ?, status = ?, cliente_id = ?, equipamento = ?, 
		servico = ?, obs_tecnicas = ?, valor_total = ?, pagamento_status = ?, 
		pagamento_forma = ?, pago_em = ?
		WHERE id = ?`

	_, err = tx.Exec(updateQuery,
		dataEmissao,
		status,
		clienteID,
		equipamento,
		strings.TrimSpace(req.Servico),
		strings.TrimSpace(req.ObsTecnicas),
		req.ValorTotal,
		pagamentoStatus,
		pagamentoForma,
		pagoEm,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update service order: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit order update: %w", err)
	}

	return d.GetOrdemByID(id)
}

// UpdatePagamento quickly marks an order as paid or updates its payment status.
func (d *DB) UpdatePagamento(id int64, forma, status string) error {
	status = strings.TrimSpace(status)
	if status == "" {
		status = "Pago"
	}
	forma = strings.TrimSpace(forma)
	if forma == "" {
		forma = "Pix"
	}

	var res sql.Result
	var err error

	if status == "Pago" {
		res, err = d.Writer.Exec(
			`UPDATE ordens_servico 
			 SET status = 'Concluído', pagamento_status = 'Pago', pagamento_forma = ?, pago_em = CURRENT_TIMESTAMP 
			 WHERE id = ?`,
			forma, id,
		)
	} else {
		res, err = d.Writer.Exec(
			`UPDATE ordens_servico 
			 SET pagamento_status = ?, pagamento_forma = ?, pago_em = NULL 
			 WHERE id = ?`,
			status, forma, id,
		)
	}

	if err != nil {
		return fmt.Errorf("failed to update payment for order #%d: %w", id, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("service order not found")
	}
	return nil
}
