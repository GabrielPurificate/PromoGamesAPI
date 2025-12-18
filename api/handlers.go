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
	// TÍTULO EM NEGRITO
	msg := fmt.Sprintf("<b>%s</b>\n\n", p.Nome)

	// LÓGICA DO PREÇO À VISTA (PIX ou NORMAL)
	if p.IsPix {
		// Se marcou o checkbox: "R$ 100,00 no PIX"
		msg += fmt.Sprintf("💰 <b>R$ %s</b> no PIX\n", p.Valor)
	} else {
		// Se NÃO marcou: "R$ 100,00" (Apenas o valor seco)
		// Se quiser escrito "à vista", troque por: "💰 <b>R$ %s</b> à vista\n"
		msg += fmt.Sprintf("💰 <b>R$ %s</b>\n", p.Valor)
	}

	// LÓGICA DO PARCELAMENTO
	if p.Parcelas > 0 {
		jurosTexto := "sem juros"
		if p.TemJuros {
			jurosTexto = "com juros"
		}
		// Ex: "💳 Ou em até 10x de R$ 50,00 sem juros"
		msg += fmt.Sprintf("💳 Ou em até %dx de R$ %s %s\n", p.Parcelas, p.ValorParcela, jurosTexto)
	}

	// CUPOM
	if p.Cupom != "" {
		msg += fmt.Sprintf("🎟 CUPOM: <code>%s</code>\n", p.Cupom)
	}

	// LINK
	msg += fmt.Sprintf("\n🔗 Link: %s\n", p.Link)

	// LOJA
	if p.Loja != "" {
		msg += fmt.Sprintf("[%s]\n", strings.ToUpper(p.Loja))
	}

	// RODAPÉ
	msg += fmt.Sprintf("\n🌐 <b>Mais ofertas em:</b> https://promogamesbr.com")

	return msg
}

func gerarSlugSimples(nome string) string {
	s := strings.ToLower(nome)
	s = strings.ReplaceAll(s, " ", "-")
	return s
}
