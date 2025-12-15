package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// Struct para o Payload do Telegram
type TelegramPhotoPayload struct {
	ChatID    string `json:"chat_id"`
	Photo     string `json:"photo"`
	Caption   string `json:"caption"`
	ParseMode string `json:"parse_mode"`
}

func EnviarParaTelegram(imagemURL, textoLegenda string) error {
	// 1. Buscamos as chaves do ambiente agora (Runtime)
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")

	// Validação de segurança básica
	if token == "" || chatID == "" {
		return fmt.Errorf("configuração do telegram ausente no .env")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendPhoto", token)

	payload := TelegramPhotoPayload{
		ChatID:    chatID,
		Photo:     imagemURL,
		Caption:   textoLegenda,
		ParseMode: "Markdown",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("erro no telegram: status %d", resp.StatusCode)
	}

	return nil
}
