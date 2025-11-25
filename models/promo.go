package models

type PromoRequest struct {
	Nome         string `json:"nome"`
	Valor        string `json:"valor"`
	Parcelamento string `json:"parcelamento"`
	Cupom        string `json:"cupom"`
	Link         string `json:"link"`
	Loja         string `json:"loja"`
}

type PromoResponse struct {
	TextoFormatado string `json:"texto_formatado"`
	ImagemUrl      string `json:"imagem_url"`
	Found          bool   `json:"found"`
}
