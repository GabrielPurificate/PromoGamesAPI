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

func formatarMensagemZap(p models.PromoRequest) string {
	msg := fmt.Sprintf("%s\n\n", p.Nome)

	msg += fmt.Sprintf("💰 R$ %s no PIX\n", p.Valor)

	if p.Parcelamento != "" {
		msg += fmt.Sprintf("💳 %s\n", p.Parcelamento)
	}

	if p.Cupom != "" {
		msg += fmt.Sprintf("🎟CUPOM: %s\n", p.Cupom)
	}

	msg += fmt.Sprintf("\n🔗 Link: %s\n", p.Link)

	if p.Loja != "" {
		msg += fmt.Sprintf("[%s]\n", strings.ToUpper(p.Loja))
	}

	msg += fmt.Sprintf("\n🌐 Mais ofertas em: https://promogamesbr.com")

	return msg
}

func gerarSlugSimples(nome string) string {
	s := strings.ToLower(nome)
	s = strings.ReplaceAll(s, " ", "-")
	return s
}
