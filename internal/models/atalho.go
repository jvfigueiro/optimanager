package models

type Atalho struct {
	ID        int64  `json:"id"`
	Titulo    string `json:"titulo"`
	Descricao string `json:"descricao"`
	Url       string `json:"url"`
	Ordem     int    `json:"ordem"`
}

