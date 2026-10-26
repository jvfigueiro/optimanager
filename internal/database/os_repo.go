package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"optimanager/internal/models"
)

// ListOrdens busca as ordens de serviço com JOIN nos clientes, aplicando filtros e busca textual.
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
		// Sem filtro de status
	case "ABERTOS":
		query += ` AND os.status != 'Concluído'`
	case "CONCLUIDOS":
		query += ` AND os.status = 'Concluído'`
	case "PENDENTE":
		query += ` AND os.status = 'Concluído' AND os.pagamento_status = 'Pendente'`
	default:
		// Se passou um status específico ("Avaliação", "Orçamento", etc.)
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
		return nil, fmt.Errorf("erro ao consultar ordens de serviço: %w", err)
	}
	defer rows.Close()

	// Obtém CNPJ da config para links de WhatsApp
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
			return nil, fmt.Errorf("erro ao mapear ordem de serviço: %w", err)
		}

		if pagoEm.Valid {
			os.PagoEm = &pagoEm.Time
		}

		os.WhatsAppURL = os.BuildWhatsAppURL(cnpj, fantasia)
		ordens = append(ordens, os)
	}

	return ordens, rows.Err()
}

// GetOrdemByID busca uma ordem de serviço completa pelo ID.
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
		return nil, fmt.Errorf("erro ao buscar ordem #%d: %w", id, err)
	}

	if pagoEm.Valid {
		os.PagoEm = &pagoEm.Time
	}

	var cnpj, fantasia string
	_ = d.Reader.QueryRow(`SELECT cnpj, fantasia FROM config WHERE id = 1`).Scan(&cnpj, &fantasia)
	os.WhatsAppURL = os.BuildWhatsAppURL(cnpj, fantasia)

	return &os, nil
}

// CreateOrdem cria uma nova Ordem de Serviço em transação com auto-cadastro de cliente e atalho VT.
func (d *DB) CreateOrdem(req *models.CreateOrdemRequest) (*models.OrdemServico, error) {
	if strings.TrimSpace(req.Servico) == "" {
		return nil, errors.New("a descrição do serviço é obrigatória")
	}

	tx, err := d.Writer.Begin()
	if err != nil {
		return nil, fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback()

	// 1. Resolução ou Auto-cadastro do Cliente
	var cliente *models.Cliente
	if req.ClienteID > 0 {
		cliente = &models.Cliente{ID: req.ClienteID}
		err = tx.QueryRow(`SELECT nome, doc, contato, desde_ano, created_at FROM clientes WHERE id = ?`, req.ClienteID).
			Scan(&cliente.Nome, &cliente.Doc, &cliente.Contato, &cliente.DesdeAno, &cliente.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("cliente informado (#%d) não encontrado: %w", req.ClienteID, err)
		}
	} else {
		if strings.TrimSpace(req.ClienteNome) == "" {
			return nil, errors.New("nome do cliente ou ID é obrigatório")
		}
		cliente, err = d.GetOrCreateCliente(tx, req.ClienteNome, req.ClienteDoc, req.ClienteContato)
		if err != nil {
			return nil, err
		}
	}

	// 2. Regra de Negócio: Atalho Visita Técnica (VT)
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

	// 3. Defaults de campos
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

	// 4. Inserção no banco
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
		return nil, fmt.Errorf("erro ao salvar ordem de serviço: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("erro ao recuperar ID gerado da OS: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("erro ao finalizar transação: %w", err)
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

// UpdateOrdem atualiza os dados de uma ordem de serviço existente.
func (d *DB) UpdateOrdem(id int64, req *models.UpdateOrdemRequest) (*models.OrdemServico, error) {
	if strings.TrimSpace(req.Servico) == "" {
		return nil, errors.New("a descrição do serviço é obrigatória")
	}

	tx, err := d.Writer.Begin()
	if err != nil {
		return nil, fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback()

	// Verifica se a OS existe
	var osExistente models.OrdemServico
	var pagoEmExistente sql.NullTime
	err = tx.QueryRow(`SELECT id, status, cliente_id, pagamento_status, pago_em FROM ordens_servico WHERE id = ?`, id).
		Scan(&osExistente.ID, &osExistente.Status, &osExistente.ClienteID, &osExistente.PagamentoStatus, &pagoEmExistente)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("ordem de serviço não encontrada")
		}
		return nil, err
	}

	// Resolução de cliente se alterado
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

	// Tratamento do pago_em
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
		return nil, fmt.Errorf("erro ao atualizar ordem de serviço: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("erro ao confirmar atualização da OS: %w", err)
	}

	return d.GetOrdemByID(id)
}

// UpdatePagamento realiza a baixa rápida de pagamento para uma OS pendente.
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
		return fmt.Errorf("erro ao atualizar pagamento da OS #%d: %w", id, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("ordem de serviço não encontrada")
	}
	return nil
}

