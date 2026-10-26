package main

import (
	"fmt"
	"log"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

func main() {
	optiBlue := &props.Color{Red: 11, Green: 86, Blue: 148}
	white := &props.Color{Red: 255, Green: 255, Blue: 255}

	cfg := config.NewBuilder().
		WithMargins(15, 15, 15).
		Build()
	m := maroto.New(cfg)

	// Header
	m.AddRows(
		row.New(12).Add(
			col.New(8).Add(
				text.New("Opti Tecnologia", props.Text{
					Style: fontstyle.Bold,
					Size:  14,
					Color: optiBlue,
				}),
			),
			col.New(4).Add(
				text.New("RELATÓRIO DE SERVIÇO", props.Text{
					Style: fontstyle.Bold,
					Size:  10,
					Align: align.Right,
					Color: optiBlue,
				}),
			),
		),
		row.New(6).Add(
			col.New(8).Add(
				text.New("CNPJ: 38.136.905/0001-66 | Fones: (55) 99127-2332", props.Text{
					Size: 8,
				}),
			),
			col.New(4).Add(
				text.New("Cód. #1000", props.Text{
					Size:  8,
					Align: align.Right,
				}),
			),
		),
		line.NewRow(3, props.Line{
			Color: optiBlue,
			Size:  1.5,
		}),
	)

	// Section title with colored background
	m.AddRows(
		row.New(7).Add(
			col.New(12).Add(
				text.New("DADOS DO ATENDIMENTO", props.Text{
					Style: fontstyle.Bold,
					Size:  9,
					Color: white,
				}),
			),
		).WithStyle(&props.Cell{
			BackgroundColor: optiBlue,
		}),
	)

	doc, err := m.Generate()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Layout gerado com sucesso: %d bytes\n", len(doc.GetBytes()))
}
