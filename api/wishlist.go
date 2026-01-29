package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/nedpals/supabase-go"
)

type TelegramUpdate struct {
	Message struct {
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
		Text string `json:"text"`
	} `json:"message"`
}

type ResultadoWishlist struct {
	TelegramID int64  `json:"telegram_id"`
	TermoBusca string `json:"termo_busca"`
}

func HandlerWebhookTelegram(client *supabase.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var update TelegramUpdate
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			http.Error(w, "Erro json", http.StatusBadRequest)
			return
		}

		chatID := update.Message.Chat.ID
		texto := strings.TrimSpace(update.Message.Text)
		textoSafe := strings.ReplaceAll(texto, "<", "")
		textoSafe = strings.ReplaceAll(textoSafe, ">", "")

		if texto == "/start" || strings.HasPrefix(texto, "/start") {
			msg := "🤖 <b>Olá! Sou o Assistente do PromoGames.</b>\n\n" +
				"Eu posso te avisar quando aquele jogo que você quer entrar em promoção!\n\n" +
				"📋 <b>COMO USAR:</b>\n" +
				"• <code>/desejo Nome do Jogo</code> → Para criar um alerta.\n" +
				"• <code>/lista</code> → Para ver seus alertas ativos.\n" +
				"• <code>/remover Nome do Jogo</code> → Para parar de receber avisos.\n\n" +
				"💡 <i>Exemplo: Digite</i> <code>/desejo God of War</code>"

			EnviarMensagemDM(chatID, msg, "")
			w.WriteHeader(http.StatusOK)
			return
		}

		if strings.HasPrefix(texto, "/desejo") {
			termo := strings.TrimSpace(strings.Replace(texto, "/desejo", "", 1))
			termo = strings.ReplaceAll(termo, "<", "")
			termo = strings.ReplaceAll(termo, ">", "")

			if len(termo) < 3 {
				EnviarMensagemDM(chatID, "❌ <b>Ops!</b> Digite o nome do jogo.\nExemplo: <code>/desejo Elden Ring</code>", "")
				w.WriteHeader(http.StatusOK)
				return
			}

			payload := map[string]interface{}{
				"telegram_id": chatID,
				"termo_busca": termo,
			}

			var result interface{}
			err := client.DB.From("alertas_usuario").Insert(payload).Execute(&result)

			if err != nil {
				log.Println("Erro ao salvar desejo:", err)
				EnviarMensagemDM(chatID, "⚠️ Erro ao salvar. Tente novamente.", "")
			} else {
				EnviarMensagemDM(chatID, fmt.Sprintf("✅ <b>Anotado!</b> Vou te avisar de promoções de: <b>%s</b>\n\n(Se comprar, use <code>/remover %s</code> para parar de receber)", termo, termo), "")
			}
			w.WriteHeader(http.StatusOK)
			return
		}

		if strings.HasPrefix(texto, "/lista") {
			var alertas []struct {
				TermoBusca string `json:"termo_busca"`
			}

			err := client.DB.From("alertas_usuario").
				Select("termo_busca").
				Filter("telegram_id", "eq", fmt.Sprintf("%d", chatID)).
				Execute(&alertas)

			if err != nil {
				EnviarMensagemDM(chatID, "Erro ao buscar sua lista.", "")
			} else if len(alertas) == 0 {
				EnviarMensagemDM(chatID, "📭 Sua lista de desejos está vazia.\nUse <code>/desejo Nome</code> para adicionar.", "")
			} else {
				msg := "📝 <b>SEUS ALERTAS ATIVOS:</b>\n(Clique no comando para copiar)\n\n"
				for _, a := range alertas {
					msg += fmt.Sprintf("🎮 <b>%s</b>\n❌ <code>/remover %s</code>\n\n", a.TermoBusca, a.TermoBusca)
				}
				EnviarMensagemDM(chatID, msg, "")
			}
			w.WriteHeader(http.StatusOK)
			return
		}

		if strings.HasPrefix(texto, "/remover") {
			termo := strings.TrimSpace(strings.Replace(texto, "/remover", "", 1))
			termo = strings.ReplaceAll(termo, "<", "")
			termo = strings.ReplaceAll(termo, ">", "")

			if len(termo) < 2 {
				EnviarMensagemDM(chatID, "❌ Digite o nome para remover.\nEx: <code>/remover God of War</code>", "")
				w.WriteHeader(http.StatusOK)
				return
			}

			var deletou bool
			err := client.DB.Rpc("remover_wishlist", map[string]interface{}{
				"p_telegram_id": chatID,
				"p_termo":       termo,
			}).Execute(&deletou)

			if err != nil {
				log.Println("ERRO RPC REMOVER:", err)
				EnviarMensagemDM(chatID, "⚠️ Erro técnico ao remover. Tente mais tarde.", "")
			} else if deletou {
				EnviarMensagemDM(chatID, fmt.Sprintf("🗑 Alerta de <b>%s</b> removido com sucesso!", termo), "")
			} else {
				EnviarMensagemDM(chatID, fmt.Sprintf("❌ Não encontrei nenhum alerta para: <b>%s</b>\n\nUse <code>/lista</code> para ver os nomes exatos.", termo), "")
			}

			w.WriteHeader(http.StatusOK)
			return
		}

		EnviarMensagemDM(chatID, "❓ Não entendi.\nDigite <code>/start</code> para ver as opções.", "")
		w.WriteHeader(http.StatusOK)
	}
}

func DispararWishlist(client *supabase.Client, tituloBusca, imagemURL, textoLegenda string) {
	go func() {
		var interessados []ResultadoWishlist

		err := client.DB.Rpc("buscar_interessados", map[string]interface{}{
			"titulo_produto": tituloBusca,
		}).Execute(&interessados)

		if err != nil {
			log.Println("Erro RPC Wishlist:", err)
			return
		}

		if len(interessados) == 0 {
			return
		}

		for _, user := range interessados {
			msgFinal := fmt.Sprintf("🚨 <b>ENCONTREI SEU PEDIDO: %s</b>\n\n", strings.ToUpper(user.TermoBusca))
			msgFinal += textoLegenda
			msgFinal += fmt.Sprintf("\n\n🗑 <i>Já comprou? Pare de receber avisos deste jogo:</i>\n<code>/remover %s</code>", user.TermoBusca)

			EnviarMensagemDM(user.TelegramID, msgFinal, imagemURL)
		}
	}()
}
