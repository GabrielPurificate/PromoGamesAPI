package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
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
		if err := validarURLTelegram(imagemURL); err != nil {
			return "", fmt.Errorf("URL de imagem inválida: %v", err)
		}
		method = "sendPhoto"
		payload["photo"] = imagemURL
		payload["caption"] = textoLegenda
	} else {
		method = "sendMessage"
		payload["text"] = textoLegenda
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/%s", token, method)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("erro ao serializar payload: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 1*1024*1024))
	if err != nil {
		return "", fmt.Errorf("erro ao ler resposta: %v", err)
	}
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

	if imagemURL != "" {
		if err := validarURLTelegram(imagemURL); err != nil {
			return fmt.Errorf("URL inválida: %v", err)
		}
		method = "sendPhoto"
		payload["photo"] = imagemURL
		payload["caption"] = texto
	} else {
		method = "sendMessage"
		payload["text"] = texto
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/%s", token, method)

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("erro ao criar payload: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, _ = io.ReadAll(io.LimitReader(resp.Body, 1*1024*1024))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("erro telegram status: %d", resp.StatusCode)
	}

	return nil
}

func validarURLTelegram(urlStr string) error {
	if urlStr == "" {
		return nil
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("URL malformada: %v", err)
	}

	if parsedURL.Scheme != "https" && parsedURL.Scheme != "http" {
		return fmt.Errorf("esquema de URL não permitido: %s", parsedURL.Scheme)
	}

	host := strings.ToLower(parsedURL.Hostname())
	proibidos := []string{"localhost", "127.0.0.1", "0.0.0.0", "169.254", "10.", "192.168"}
	for _, p := range proibidos {
		if strings.HasPrefix(host, p) || strings.Contains(host, p) {
			return fmt.Errorf("acesso a rede privada bloqueado")
		}
	}

	return nil
}
