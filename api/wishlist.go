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
		texto := update.Message.Text

		if strings.HasPrefix(texto, "/desejo") {
			termo := strings.TrimSpace(strings.Replace(texto, "/desejo", "", 1))

			if len(termo) < 3 {
				EnviarMensagemDM(chatID, "❌ *Ops!* Digite o nome do jogo.\nExemplo: `/desejo God of War`")
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
				EnviarMensagemDM(chatID, fmt.Sprintf("✅ *Anotado!* Vou te avisar assim que aparecer promoção de: *%s*", termo))
			}
		}

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
