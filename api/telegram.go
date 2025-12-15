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
