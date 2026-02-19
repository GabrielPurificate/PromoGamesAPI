package models

type PromoRequest struct {
	Nome  string `json:"nome"`
	Valor string `json:"valor"`
	Loja  string `json:"loja"`
	Link  string `json:"link"`
	Cupom string `json:"cupom"`

	Parcelas     int    `json:"parcelas"`
	ValorParcela string `json:"valor_parcela"`
	TemJuros     bool   `json:"tem_juros"`

	IsPix bool `json:"is_pix"`
}

type PromoResponse struct {
	TextoFormatado string `json:"texto_formatado"`
	ImageUrl       string `json:"image_url"`
	Found          bool   `json:"found"`
}

type JogoBusca struct {
	Slug     string `json:"slug"`
	ImageUrl string `json:"image_url"`
}

type CupomRecente struct {
	Cupom   string `json:"cupom"`
	UsadoEm string `json:"usado_em"`
}

type JogoPopular struct {
	JogoSlug    string `json:"jogo_slug"`
	JogoNome    string `json:"jogo_nome"`
	TotalEnvios int    `json:"total_envios"`
	UltimoEnvio string `json:"ultimo_envio"`
}
