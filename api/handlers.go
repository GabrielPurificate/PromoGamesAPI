package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/GabrielPurificate/PromoGamesAPI/models"
	"github.com/nedpals/supabase-go"
)

func HandlerGerarPreview(client *supabase.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1*1024*1024)

		var req models.PromoRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Erro ao decodificar requisição", http.StatusBadRequest)
			return
		}

		if len(req.Nome) > 200 || len(req.Link) > 500 {
			http.Error(w, "Dados inválidos: tamanho excedido", http.StatusBadRequest)
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

		if req.Cupom != "" {
			go func() {
				defer func() {
					if r := recover(); r != nil {
						fmt.Printf("Erro ao salvar cupom: %v\n", r)
					}
				}()
				payload := map[string]interface{}{
					"cupom": req.Cupom,
				}
				var result interface{}
				errCupom := client.DB.From("cupons_recentes").Insert(payload).Execute(&result)
				if errCupom != nil {
					fmt.Printf("Erro ao salvar cupom recente: %v\n", errCupom)
				}
			}()
		}
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

		r.Body = http.MaxBytesReader(w, r.Body, 2*1024*1024)

		var dados struct {
			Texto  string `json:"texto"`
			Imagem string `json:"imagem"`
		}
		if err := json.NewDecoder(r.Body).Decode(&dados); err != nil {
			http.Error(w, "Erro JSON", http.StatusBadRequest)
			return
		}

		if len(dados.Texto) > 4096 || len(dados.Imagem) > 500 {
			http.Error(w, "Dados muito grandes", http.StatusBadRequest)
			return
		}

		respostaTelegram, err := EnviarParaTelegram(dados.Imagem, dados.Texto)

		if err != nil {
			fmt.Printf("ERRO HANDLER TELEGRAM: %v || RESPOSTA: %s\n", err, respostaTelegram)
			http.Error(w, "Erro interno ao processar envio para o Telegram.", http.StatusInternalServerError)
			return
		}

		go func() {
			channelID := os.Getenv("WHATSAPP_CHANNEL_ID")
			if channelID == "" {
				fmt.Println("⚠️ WhatsApp pulado: WHATSAPP_CHANNEL_ID não configurado.")
				return
			}

			textoZap := converterHTMLparaMarkdown(dados.Texto)

			errZap := EnviarMensagemCanal(channelID, textoZap, dados.Imagem)
			if errZap != nil {
				fmt.Printf("❌ Erro ao enviar para WhatsApp: %v\n", errZap)
			} else {
				fmt.Println("✅ Enviado para WhatsApp com sucesso!")
			}
		}()

		titulo, link := extrairDadosDoTexto(dados.Texto)
		imagemCopy := dados.Imagem
		textoCopy := dados.Texto

		if titulo != "" && link != "" {
			go func() {
				defer func() {
					if r := recover(); r != nil {
						fmt.Printf("Erro em DispararWishlist: %v\n", r)
					}
				}()
				DispararWishlist(client, titulo, imagemCopy, textoCopy)
			}()
		}

		if titulo != "" {
			go func() {
				defer func() {
					if r := recover(); r != nil {
						fmt.Printf("Erro ao incrementar contagem: %v\n", r)
					}
				}()
				slugJogo := gerarSlugSimples(titulo)
				var result interface{}
				errContagem := client.DB.Rpc("incrementar_contagem_promocao", map[string]interface{}{
					"p_slug": slugJogo,
					"p_nome": titulo,
				}).Execute(&result)
				if errContagem != nil {
					fmt.Printf("Erro ao incrementar contagem de promoção: %v\n", errContagem)
				}
			}()
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "sucesso",
			"msg":    "Promoção enviada! (Telegram: OK | Zap: Fila | Wishlist: Fila)",
		})
	}
}

func converterHTMLparaMarkdown(textoHTML string) string {
	t := textoHTML
	t = strings.ReplaceAll(t, "<b>", "*")
	t = strings.ReplaceAll(t, "</b>", "*")
	t = strings.ReplaceAll(t, "<strong>", "*")
	t = strings.ReplaceAll(t, "</strong>", "*")
	t = strings.ReplaceAll(t, "<i>", "_")
	t = strings.ReplaceAll(t, "</i>", "_")
	t = strings.ReplaceAll(t, "<em>", "_")
	t = strings.ReplaceAll(t, "</em>", "_")
	t = strings.ReplaceAll(t, "<code>", "```")
	t = strings.ReplaceAll(t, "</code>", "```")

	return t
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

	msg += "\n🌐 <b>Mais ofertas em:</b> https://promogamesbr.com"

	return msg
}

func gerarSlugSimples(nome string) string {
	s := strings.ToLower(nome)
	reg := regexp.MustCompile(`[^a-z0-9\s-]+`)
	s = reg.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, " ", "-")
	return s
}

func HandlerBuscarJogos(client *supabase.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		termo := strings.TrimSpace(r.URL.Query().Get("q"))
		if len(termo) < 2 || len(termo) > 100 {
			http.Error(w, "Parâmetro 'q' deve ter entre 2 e 100 caracteres", http.StatusBadRequest)
			return
		}

		var resultados []models.JogoBusca

		slugBusca := strings.ToLower(strings.TrimSpace(termo))
		slugBusca = strings.ReplaceAll(slugBusca, " ", "*")

		err := client.DB.From("Jogos").
			Select("slug", "image_url").
			Limit(10).
			Ilike("slug", "*"+slugBusca+"*").
			Execute(&resultados)

		if err != nil {
			fmt.Printf("Erro ao buscar jogos: %v\n", err)
			http.Error(w, "Erro ao buscar jogos", http.StatusInternalServerError)
			return
		}

		if resultados == nil {
			resultados = []models.JogoBusca{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resultados)
	}
}

func HandlerCuponsRecentes(client *supabase.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		var resultados []models.CupomRecente

		err := client.DB.From("cupons_recentes").
			Select("cupom", "usado_em").
			OrderBy("usado_em", "desc").
			Limit(20).
			Execute(&resultados)

		if err != nil {
			fmt.Printf("Erro ao buscar cupons recentes: %v\n", err)
			http.Error(w, "Erro ao buscar cupons recentes", http.StatusInternalServerError)
			return
		}

		vistos := make(map[string]bool)
		var unicos []models.CupomRecente
		for _, c := range resultados {
			cupomLower := strings.ToLower(strings.TrimSpace(c.Cupom))
			if !vistos[cupomLower] {
				vistos[cupomLower] = true
				unicos = append(unicos, c)
				if len(unicos) >= 3 {
					break
				}
			}
		}

		if unicos == nil {
			unicos = []models.CupomRecente{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(unicos)
	}
}

func HandlerJogosPopulares(client *supabase.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		var resultados []models.JogoPopular

		err := client.DB.From("contagem_promocoes").
			Select("jogo_slug", "jogo_nome", "total_envios", "ultimo_envio").
			OrderBy("total_envios", "desc").
			Limit(10).
			Execute(&resultados)

		if err != nil {
			fmt.Printf("Erro ao buscar jogos populares: %v\n", err)
			http.Error(w, "Erro ao buscar jogos populares", http.StatusInternalServerError)
			return
		}

		if resultados == nil {
			resultados = []models.JogoPopular{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resultados)
	}
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
