package pdf

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/pagesize"
	"github.com/johnfercher/maroto/v2/pkg/props"

	"optimanager/internal/models"
)

// HexToColor converte string hex ("#0B5694") para *props.Color do Maroto
func HexToColor(hex string) *props.Color {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return &props.Color{Red: 11, Green: 86, Blue: 148} // Padrão Opti Blue
	}
	r, err1 := strconv.ParseUint(hex[0:2], 16, 8)
	g, err2 := strconv.ParseUint(hex[2:4], 16, 8)
	b, err3 := strconv.ParseUint(hex[4:6], 16, 8)
	if err1 != nil || err2 != nil || err3 != nil {
		return &props.Color{Red: 11, Green: 86, Blue: 148}
	}
	return &props.Color{Red: int(r), Green: int(g), Blue: int(b)}
}

// FormatBRL formata valores para padrão moeda brasileira
func FormatBRL(v float64) string {
	return fmt.Sprintf("R$ %.2f", v)
}

// FormatDate formata YYYY-MM-DD para DD/MM/YYYY
func FormatDate(d string) string {
	parts := strings.Split(strings.Split(d, "T")[0], "-")
	if len(parts) == 3 {
		return fmt.Sprintf("%s/%s/%s", parts[2], parts[1], parts[0])
	}
	return d
}

// GenerateDocument cria o PDF nativo A4 de Relatório, Orçamento ou Recibo usando Maroto v2.
func GenerateDocument(cfg *models.Config, os *models.OrdemServico, docType string) ([]byte, error) {
	primaryColor := HexToColor(cfg.CorPrimaria)
	white := &props.Color{Red: 255, Green: 255, Blue: 255}
	slate700 := &props.Color{Red: 51, Green: 65, Blue: 85}
	slate500 := &props.Color{Red: 100, Green: 116, Blue: 139}
	bgLight := &props.Color{Red: 248, Green: 250, Blue: 252}

	marotoCfg := config.NewBuilder().
		WithPageSize(pagesize.A4).
		WithLeftMargin(15).
		WithRightMargin(15).
		WithTopMargin(15).
		WithBottomMargin(15).
		WithSequentialMode().
		Build()

	m := maroto.New(marotoCfg)

	docType = strings.ToLower(strings.TrimSpace(docType))
	var tituloDoc string
	switch docType {
	case "orcamento":
		tituloDoc = "ORÇAMENTO TÉCNICO"
	case "recibo":
		tituloDoc = "RECIBO DE PAGAMENTO"
	default:
		docType = "relatorio"
		tituloDoc = "RELATÓRIO DE SERVIÇO"
	}

	dataEmissao := FormatDate(os.DataEmissao)
	if dataEmissao == "" {
		dataEmissao = time.Now().Format("02/01/2006")
	}

	// 1. Cabeçalho Institucional
	m.AddRows(
		row.New(9).Add(
			col.New(8).Add(
				text.New(cfg.Fantasia, props.Text{
					Style: fontstyle.Bold,
					Size:  15,
					Color: primaryColor,
				}),
			),
			col.New(4).Add(
				text.New(tituloDoc, props.Text{
					Style: fontstyle.Bold,
					Size:  11,
					Align: align.Right,
					Color: primaryColor,
				}),
			),
		),
		row.New(5).Add(
			col.New(8).Add(
				text.New(fmt.Sprintf("CNPJ: %s | Fones: %s", cfg.CNPJ, cfg.Fones), props.Text{
					Size:  8,
					Color: slate700,
				}),
			),
			col.New(4).Add(
				text.New(fmt.Sprintf("Cód. #%d", os.ID), props.Text{
					Style: fontstyle.Bold,
					Size:  9,
					Align: align.Right,
					Color: slate700,
				}),
			),
		),
		row.New(4).Add(
			col.New(12).Add(
				text.New(fmt.Sprintf("Responsável: %s | %s", cfg.Responsavel, cfg.Cidade), props.Text{
					Size:  8,
					Color: slate500,
				}),
			),
		),
		line.NewRow(3, props.Line{
			Color:     primaryColor,
			Thickness: 1.2,
		}),
		row.New(3).Add(col.New(12).Add(text.New("", props.Text{}))), // Espaçador
	)

	// 2. Miolo do Documento conforme o Tipo
	switch docType {
	case "relatorio":
		// Dados do Atendimento
		m.AddRows(
			row.New(6).Add(
				col.New(12).Add(
					text.New("  DADOS DO ATENDIMENTO", props.Text{
						Style: fontstyle.Bold,
						Size:  9,
						Color: white,
						Top:   1,
					}),
				),
			).WithStyle(&props.Cell{BackgroundColor: primaryColor}),
			row.New(1).Add(col.New(12).Add(text.New("", props.Text{}))),
			row.New(5).Add(
				col.New(6).Add(text.New(fmt.Sprintf("Data de Emissão: %s", dataEmissao), props.Text{Size: 8.5})),
				col.New(6).Add(text.New(fmt.Sprintf("Cidade: %s", cfg.Cidade), props.Text{Size: 8.5})),
			),
			row.New(5).Add(
				col.New(6).Add(text.New(fmt.Sprintf("Cliente: %s", os.ClienteNome), props.Text{Style: fontstyle.Bold, Size: 8.5})),
				col.New(6).Add(text.New(fmt.Sprintf("CPF/CNPJ: %s", defaultStr(os.ClienteDoc, "N/A")), props.Text{Size: 8.5})),
			),
			row.New(5).Add(
				col.New(12).Add(text.New(fmt.Sprintf("Equipamento: %s", os.Equipamento), props.Text{Size: 8.5})),
			),
			row.New(4).Add(col.New(12).Add(text.New("", props.Text{}))),
			// Descrição dos Serviços
			row.New(6).Add(
				col.New(12).Add(
					text.New("  DESCRIÇÃO DOS SERVIÇOS REALIZADOS", props.Text{
						Style: fontstyle.Bold,
						Size:  9,
						Color: white,
						Top:   1,
					}),
				),
			).WithStyle(&props.Cell{BackgroundColor: primaryColor}),
			row.New(2).Add(col.New(12).Add(text.New("", props.Text{}))),
			row.New(28).Add(
				col.New(12).Add(
					text.New(os.Servico, props.Text{
						Size:  8.5,
						Color: slate700,
					}),
				),
			).WithStyle(&props.Cell{BackgroundColor: bgLight}),
		)

		if strings.TrimSpace(os.ObsTecnicas) != "" {
			m.AddRows(
				row.New(3).Add(col.New(12).Add(text.New("", props.Text{}))),
				row.New(10).Add(
					col.New(12).Add(
						text.New(fmt.Sprintf("Parâmetros Técnicos / Observações: %s", os.ObsTecnicas), props.Text{
							Style: fontstyle.Bold,
							Size:  8,
							Color: primaryColor,
						}),
					),
				),
			)
		}

		if strings.TrimSpace(cfg.TermoGarantia) != "" {
			m.AddRows(
				row.New(3).Add(col.New(12).Add(text.New("", props.Text{}))),
				row.New(8).Add(
					col.New(12).Add(
						text.New(fmt.Sprintf("Garantia: %s", cfg.TermoGarantia), props.Text{
							Size:  7.5,
							Color: slate500,
						}),
					),
				),
			)
		}

		// Valor
		m.AddRows(
			row.New(4).Add(col.New(12).Add(text.New("", props.Text{}))),
			row.New(7).Add(
				col.New(12).Add(
					text.New(fmt.Sprintf("Valor Total do Atendimento: %s", FormatBRL(os.ValorTotal)), props.Text{
						Style: fontstyle.Bold,
						Size:  10,
						Align: align.Right,
						Color: primaryColor,
					}),
				),
			),
		)

	case "orcamento":
		m.AddRows(
			row.New(5).Add(
				col.New(6).Add(text.New(fmt.Sprintf("Cliente: %s", os.ClienteNome), props.Text{Style: fontstyle.Bold, Size: 8.5})),
				col.New(6).Add(text.New(fmt.Sprintf("Data: %s (Validade: 30 dias)", dataEmissao), props.Text{Size: 8.5})),
			),
			row.New(5).Add(
				col.New(6).Add(text.New(fmt.Sprintf("Equipamento: %s", os.Equipamento), props.Text{Size: 8.5})),
				col.New(6).Add(text.New(fmt.Sprintf("Cidade: %s", cfg.Cidade), props.Text{Size: 8.5})),
			),
			row.New(4).Add(col.New(12).Add(text.New("", props.Text{}))),
			row.New(6).Add(
				col.New(12).Add(
					text.New("  DESCRITIVO DO ORÇAMENTO", props.Text{
						Style: fontstyle.Bold,
						Size:  9,
						Color: white,
						Top:   1,
					}),
				),
			).WithStyle(&props.Cell{BackgroundColor: primaryColor}),
			row.New(2).Add(col.New(12).Add(text.New("", props.Text{}))),
			row.New(30).Add(
				col.New(12).Add(
					text.New(os.Servico, props.Text{
						Size:  8.5,
						Color: slate700,
					}),
				),
			).WithStyle(&props.Cell{BackgroundColor: bgLight}),
		)

		if strings.TrimSpace(os.ObsTecnicas) != "" {
			m.AddRows(
				row.New(3).Add(col.New(12).Add(text.New("", props.Text{}))),
				row.New(10).Add(
					col.New(12).Add(
						text.New(fmt.Sprintf("Observações Adicionais: %s", os.ObsTecnicas), props.Text{
							Size:  8,
							Color: slate700,
						}),
					),
				),
			)
		}

		m.AddRows(
			row.New(4).Add(col.New(12).Add(text.New("", props.Text{}))),
			row.New(8).Add(
				col.New(12).Add(
					text.New(fmt.Sprintf("VALOR TOTAL: %s", FormatBRL(os.ValorTotal)), props.Text{
						Style: fontstyle.Bold,
						Size:  12,
						Align: align.Right,
						Color: primaryColor,
					}),
				),
			),
		)

	case "recibo":
		docCli := os.ClienteDoc
		if docCli != "" {
			docCli = fmt.Sprintf(" (CPF/CNPJ: %s)", docCli)
		}

		textoRecibo := fmt.Sprintf(
			"Eu, %s, inscrito sob o CNPJ %s, recebi de %s%s, a importância de %s (%s), referente aos serviços de: %s. E para clareza e fins de direito, firmo o presente recibo de quitação.",
			cfg.RazaoSocial,
			cfg.CNPJ,
			os.ClienteNome,
			docCli,
			FormatBRL(os.ValorTotal),
			os.PagamentoForma,
			os.Servico,
		)

		m.AddRows(
			row.New(10).Add(
				col.New(6).Add(
					text.New(fmt.Sprintf("Data de Liquidação: %s", dataEmissao), props.Text{
						Style: fontstyle.Bold,
						Size:  9,
						Color: slate700,
					}),
				),
				col.New(6).Add(
					text.New(fmt.Sprintf("Valor Recebido: %s", FormatBRL(os.ValorTotal)), props.Text{
						Style: fontstyle.Bold,
						Size:  12,
						Align: align.Right,
						Color: primaryColor,
					}),
				),
			).WithStyle(&props.Cell{BackgroundColor: bgLight}),
			row.New(6).Add(col.New(12).Add(text.New("", props.Text{}))),
			row.New(35).Add(
				col.New(12).Add(
					text.New(textoRecibo, props.Text{
						Size:  9.5,
						Color: slate700,
					}),
				),
			),
			row.New(8).Add(
				col.New(12).Add(
					text.New(fmt.Sprintf("%s, %s.", cfg.Cidade, dataEmissao), props.Text{
						Style: fontstyle.Italic,
						Size:  8.5,
						Align: align.Right,
						Color: slate500,
					}),
				),
			),
		)
	}

	// 3. Rodapé com Assinaturas
	m.AddRows(
		row.New(15).Add(col.New(12).Add(text.New("", props.Text{}))), // Espaçador
		row.New(1).Add(
			col.New(5).Add(line.New(props.Line{Color: slate500, Thickness: 0.8})),
			col.New(2).Add(text.New("", props.Text{})),
			col.New(5).Add(line.New(props.Line{Color: slate500, Thickness: 0.8})),
		),
		row.New(5).Add(
			col.New(5).Add(
				text.New(fmt.Sprintf("%s\n%s", cfg.Responsavel, cfg.Fantasia), props.Text{
					Size:  7.5,
					Align: align.Center,
					Color: slate700,
				}),
			),
			col.New(2).Add(text.New("", props.Text{})),
			col.New(5).Add(
				text.New(fmt.Sprintf("%s\nCliente / Tomador", os.ClienteNome), props.Text{
					Size:  7.5,
					Align: align.Center,
					Color: slate700,
				}),
			),
		),
		row.New(8).Add(col.New(12).Add(text.New("", props.Text{}))),
		line.NewRow(1, props.Line{Color: slate500, Thickness: 0.4}),
		row.New(4).Add(
			col.New(12).Add(
				text.New(fmt.Sprintf("Emitido via OptiManager 3.0 • %s • %s", cfg.Fantasia, cfg.Cidade), props.Text{
					Style: fontstyle.Bold,
					Size:  7,
					Align: align.Center,
					Color: primaryColor,
				}),
			),
		),
	)

	doc, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar PDF com Maroto: %w", err)
	}

	return doc.GetBytes(), nil
}

func defaultStr(val, def string) string {
	if strings.TrimSpace(val) == "" {
		return def
	}
	return val
}

