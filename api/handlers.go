package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
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
		fmt.Printf("🔍 Buscando por slug: [%s]\n", slugBusca)

		var resultados []struct {
			ImageUrl string `json:"image_url"`
		}

		err := client.DB.From("Jogos").
			Select("image_url").
			Limit(1).
			Filter("slug", "eq", slugBusca).
			Execute(&resultados)

		imagemFinal := ""
		achou := false

		if err == nil && len(resultados) > 0 {
			imagemFinal = resultados[0].ImageUrl
			achou = true
		}

		texto := formatarMensagemZap(req)

		resp := models.PromoResponse{
			TextoFormatado: texto,
			ImageUrl:       imagemFinal,
			Found:          achou,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func HandlerPing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"msg":    "pong",
	})
}

func HandlerEnviarTelegram(client *supabase.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

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

		go func() {
			titulo, link := extrairDadosDoTexto(dados.Texto)

			if titulo != "" && link != "" {
				fmt.Printf("🎯 Wishlist Trigger: Buscando interessados em [%s]\n", titulo)
				DispararWishlist(client, titulo, link)
			} else {
				fmt.Println("⚠️ Não foi possível extrair título/link para Wishlist")
			}
		}()

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "sucesso",
			"msg":    "Telegram respondeu: " + respostaTelegram,
		})
	}
}

func formatarMensagemZap(p models.PromoRequest) string {
	msg := fmt.Sprintf("<b>%s</b>\n\n", p.Nome)

	if p.IsPix {
		msg += fmt.Sprintf("💰 <b>R$ %s</b> no PIX\n", p.Valor)
	} else {
		msg += fmt.Sprintf("💰 <b>R$ %s</b>\n", p.Valor)
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
		msg += fmt.Sprintf("[%s]\n", strings.ToUpper(p.Loja))
	}

	msg += fmt.Sprintf("\n🌐 <b>Mais ofertas em:</b> https://promogamesbr.com")

	return msg
}

func gerarSlugSimples(nome string) string {
	s := strings.ToLower(nome)
	reg, _ := regexp.Compile("[^a-z0-9\\s-]+")
	s = reg.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, " ", "-")
	return s
}

func extrairDadosDoTexto(textoHTML string) (string, string) {
	linhas := strings.Split(textoHTML, "\n")
	titulo := ""

	if len(linhas) > 0 {
		titulo = linhas[0]
		titulo = strings.ReplaceAll(titulo, "<b>", "")
		titulo = strings.ReplaceAll(titulo, "</b>", "")
		titulo = strings.TrimSpace(titulo)
	}

	reLink := regexp.MustCompile(`https?://[^\s]+`)
	link := reLink.FindString(textoHTML)

	fmt.Printf("DEBUG EXTRAÇÃO -> Titulo: [%s] | Link: [%s]\n", titulo, link)

	return titulo, link
}
