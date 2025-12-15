package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/GabrielPurificate/PromoGamesAPI/models"

	"github.com/nedpals/supabase-go"
)

func HandlerGerarPreview(client *supabase.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req models.PromoRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Erro ao decodificar requisição: "+err.Error(), http.StatusBadRequest)
			return
		}

		slugBusca := gerarSlugSimples(req.Nome)

		var resultados []struct {
			ImagemUrl string `json:"imagem_url"`
		}

		err := client.DB.From("jogos").
			Select("image_url").
			Limit(1).
			Filter("slug", "ilike", "%"+slugBusca+"%").
			Execute(&resultados)

		imagemFinal := ""
		achou := false

		if err == nil && len(resultados) > 0 {
			imagemFinal = resultados[0].ImagemUrl
			achou = true
		} else {
			imagemFinal = ""
			achou = false
		}

		texto := formatarMensagemZap(req)

		resp := models.PromoResponse{
			TextoFormatado: texto,
			ImagemUrl:      imagemFinal,
			Found:          achou,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func HandlerPing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	resp := map[string]string{
		"status": "ok",
		"msg":    "pong",
	}

	json.NewEncoder(w).Encode(resp)
}

func HandlerEnviarTelegram(w http.ResponseWriter, r *http.Request) {

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var dados struct {
		Texto  string `json:"texto"`
		Imagem string `json:"imagem"`
	}
	if err := json.NewDecoder(r.Body).Decode(&dados); err != nil {
		http.Error(w, "Erro JSON", http.StatusBadRequest)
		return
	}

	respostaTelegram, err := EnviarParaTelegram(dados.Imagem, dados.Texto)

	if err != nil {
		http.Error(w, "FALHA: "+err.Error()+" || RESPOSTA TELEGRAM: "+respostaTelegram, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "sucesso",
		"msg":    "Telegram respondeu: " + respostaTelegram,
	})
}

func formatarMensagemZap(p models.PromoRequest) string {
	msg := fmt.Sprintf("🔥 <b>%s</b>\n\n", p.Nome)

	tipoPag := "no PIX"
	if p.TipoPagamento != "" {
		tipoPag = p.TipoPagamento
	}

	if p.TipoPagamento == "NORMAL" {
		msg += fmt.Sprintf("💰 <b>R$ %s</b>\n", p.Valor)
	} else {
		msg += fmt.Sprintf("💰 <b>R$ %s</b> %s\n", p.Valor, tipoPag)
	}

	if p.Parcelas > 0 {
		jurosTexto := "sem juros"
		if p.TemJuros {
			jurosTexto = "com juros"
		}
		msg += fmt.Sprintf("💳 Ou em até %dx de R$ %s %s\n", p.Parcelas, p.ValorParcela, jurosTexto)
	}

	if p.Cupom != "" {
		msg += fmt.Sprintf("🎟 CUPOM: <code>%s</code>\n", p.Cupom)
	}

	msg += fmt.Sprintf("\n🔗 Link: %s\n", p.Link)

	if p.Loja != "" {
		msg += fmt.Sprintf("\n🏪 Loja: %s\n", strings.ToUpper(p.Loja))
	}

	msg += fmt.Sprintf("\n🌐 <b>Mais ofertas em:</b> https://promogamesbr.com")

	return msg
}

func gerarSlugSimples(nome string) string {
	s := strings.ToLower(nome)
	s = strings.ReplaceAll(s, " ", "-")
	return s
}
