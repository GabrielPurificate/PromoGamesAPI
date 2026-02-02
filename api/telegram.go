package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func EnviarParaTelegram(imagemURL, textoLegenda string) (string, error) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")

	if token == "" || chatID == "" {
		return "", fmt.Errorf("variáveis de ambiente vazias (verifique o Render)")
	}

	payload := make(map[string]interface{})
	payload["chat_id"] = chatID
	payload["parse_mode"] = "HTML"

	/* --- COMENTADO TEMPORARIAMENTE ---
	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]string{
			{
				{
					"text": "🔔 Criar Alerta de Preço",
					"url":  "https://t.me/PromoGamesBR_bot?start=vim_do_canal",
				},
			},
		},
	}

	payload["reply_markup"] = keyboard
	----------------------------------- */

	var method string
	if imagemURL != "" {
		method = "sendPhoto"
		payload["photo"] = imagemURL
		payload["caption"] = textoLegenda
	} else {
		method = "sendMessage"
		payload["text"] = textoLegenda
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/%s", token, method)
	jsonPayload, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	respostaTelegram := string(bodyBytes)

	if resp.StatusCode != http.StatusOK {
		return respostaTelegram, fmt.Errorf("status %d", resp.StatusCode)
	}

	return respostaTelegram, nil
}

func EnviarMensagemDM(chatID int64, texto string, imagemURL string) error {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return fmt.Errorf("token do bot não configurado")
	}

	payload := make(map[string]interface{})
	payload["chat_id"] = chatID
	payload["parse_mode"] = "HTML"

	var method string
	var url string

	if imagemURL != "" {
		method = "sendPhoto"
		payload["photo"] = imagemURL
		payload["caption"] = texto
	} else {
		method = "sendMessage"
		payload["text"] = texto
	}

	url = fmt.Sprintf("https://api.telegram.org/bot%s/%s", token, method)

	jsonPayload, _ := json.Marshal(payload)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("erro telegram status: %d", resp.StatusCode)
	}

	return nil
}
