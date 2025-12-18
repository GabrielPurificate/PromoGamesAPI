package models

type PromoRequest struct {
	Nome  string `json:"nome"`
	Valor string `json:"valor"`
	Loja  string `json:"loja"`
	Link  string `json:"link"`
	Cupom string `json:"cupom"`

	// Parcelamento
	Parcelas     int    `json:"parcelas"`
	ValorParcela string `json:"valor_parcela"`
	TemJuros     bool   `json:"tem_juros"`

	// --- NOVO CAMPO ---
	IsPix bool `json:"is_pix"`
}

type PromoResponse struct {
	TextoFormatado string `json:"texto_formatado"`
	ImagemUrl      string `json:"imagem_url"`
	Found          bool   `json:"found"`
}
