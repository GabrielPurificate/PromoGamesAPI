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

func HandlerWebhookTelegram(client *supabase.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var update TelegramUpdate
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			http.Error(w, "Erro json", http.StatusBadRequest)
			return
		}

		chatID := update.Message.Chat.ID
		texto := strings.TrimSpace(update.Message.Text)

		if texto == "/start" || strings.HasPrefix(texto, "/start") {
			msg := "🤖 *Olá! Sou o Assistente do PromoGames.*\n\n" +
				"Eu posso te avisar quando aquele jogo que você quer entrar em promoção!\n\n" +
				"📋 *COMO USAR:*\n" +
				"• `/desejo Nome do Jogo` → Para criar um alerta.\n" +
				"• `/lista` → Para ver seus alertas ativos.\n" +
				"• `/remover Nome do Jogo` → Para parar de receber avisos.\n\n" +
				"💡 _Exemplo: Digite_ `/desejo God of War`"

			EnviarMensagemDM(chatID, msg)
			w.WriteHeader(http.StatusOK)
			return
		}

		if strings.HasPrefix(texto, "/desejo") {
			termo := strings.TrimSpace(strings.Replace(texto, "/desejo", "", 1))

			if len(termo) < 3 {
				EnviarMensagemDM(chatID, "❌ *Ops!* Digite o nome do jogo.\nExemplo: `/desejo Elden Ring`")
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
				EnviarMensagemDM(chatID, "⚠️ Erro ao salvar. Tente novamente.")
			} else {
				EnviarMensagemDM(chatID, fmt.Sprintf("✅ *Anotado!* Vou te avisar de promoções de: *%s*\n\n(Se comprar, use `/remover %s` para parar de receber)", termo, termo))
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
				EnviarMensagemDM(chatID, "Erro ao buscar sua lista.")
			} else if len(alertas) == 0 {
				EnviarMensagemDM(chatID, "📭 Sua lista de desejos está vazia.\nUse `/desejo Nome` para adicionar.")
			} else {
				msg := "📝 *SEUS ALERTAS ATIVOS:*\n\n"
				for _, a := range alertas {
					msg += fmt.Sprintf("• %s\n", a.TermoBusca)
				}
				msg += "\n🗑 Para apagar, use `/remover NomeDoJogo`"
				EnviarMensagemDM(chatID, msg)
			}
			w.WriteHeader(http.StatusOK)
			return
		}

		if strings.HasPrefix(texto, "/remover") {
			termo := strings.TrimSpace(strings.Replace(texto, "/remover", "", 1))

			if len(termo) < 2 {
				EnviarMensagemDM(chatID, "❌ Digite o nome para remover.\nEx: `/remover God of War`")
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
				EnviarMensagemDM(chatID, "⚠️ Erro técnico ao remover. Tente mais tarde.")
			} else if deletou {
				EnviarMensagemDM(chatID, fmt.Sprintf("🗑 Alerta de *%s* removido com sucesso!", termo))
			} else {
				EnviarMensagemDM(chatID, fmt.Sprintf("❌ Não encontrei nenhum alerta para: *%s*\n\nUse `/lista` para ver os nomes exatos.", termo))
			}

			w.WriteHeader(http.StatusOK)
			return
		}

		EnviarMensagemDM(chatID, "❓ Não entendi.\nDigite `/start` para ver as opções.")
		w.WriteHeader(http.StatusOK)
	}
}

type ResultadoWishlist struct {
	TelegramID int64  `json:"telegram_id"`
	TermoBusca string `json:"termo_busca"`
}

func DispararWishlist(client *supabase.Client, tituloJogo, linkJogo string) {
	go func() {
		var interessados []ResultadoWishlist

		err := client.DB.Rpc("buscar_interessados", map[string]interface{}{
			"titulo_produto": tituloJogo,
		}).Execute(&interessados)

		if err != nil {
			log.Println("Erro RPC Wishlist:", err)
			return
		}

		if len(interessados) == 0 {
			return
		}

		log.Printf("Wishlist: Encontrados %d interessados em '%s'", len(interessados), tituloJogo)

		for _, user := range interessados {
			msg := fmt.Sprintf("🚨 *ACHEI SEU PEDIDO!*\n\nVocê pediu *%s* e acabou de sair: *%s*\n\n👉 [Corre pra ver](%s)", user.TermoBusca, tituloJogo, linkJogo)
			EnviarMensagemDM(user.TelegramID, msg)
		}
	}()
}
