package models

type PromoRequest struct {
	Nome          string `json:"nome"`
	Valor         string `json:"valor"`
	TipoPagamento string `json:"tipo_pagamento"` // Ex: "PIX", "BOLETO", "AVISTA" (padrão PIX)

	// Parcelamento
	Parcelas     int    `json:"parcelas"`      // Ex: 10
	ValorParcela string `json:"valor_parcela"` // Ex: "484,25"
	TemJuros     bool   `json:"tem_juros"`     // true ou false

	Cupom string `json:"cupom"`
	Link  string `json:"link"`
	Loja  string `json:"loja"`
}

type PromoResponse struct {
	TextoFormatado string `json:"texto_formatado"`
	ImagemUrl      string `json:"imagem_url"`
	Found          bool   `json:"found"`
}
