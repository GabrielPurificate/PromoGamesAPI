package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/skip2/go-qrcode"
)

func HandlerGetQR(w http.ResponseWriter, r *http.Request) {
	waClientMutex.RLock()
	if WAClient == nil {
		waClientMutex.RUnlock()
		http.Error(w, "Cliente WhatsApp não iniciado", http.StatusInternalServerError)
		return
	}

	if WAClient.Store.ID != nil {
		waClientMutex.RUnlock()
		w.Write([]byte("WhatsApp já está conectado."))
		return
	}
	waClientMutex.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	qrChan, err := WAClient.GetQRChannel(ctx)
	if err != nil {
		http.Error(w, "Erro ao obter canal QR: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = WAClient.Connect()
	if err != nil {
		http.Error(w, "Erro ao conectar: "+err.Error(), 500)
		return
	}

	for evt := range qrChan {
		select {
		case <-ctx.Done():
			http.Error(w, "Timeout ao gerar QR code", http.StatusRequestTimeout)
			return
		default:
		}

		if evt.Event == "code" {
			png, err := qrcode.Encode(evt.Code, qrcode.Medium, 256)
			if err != nil {
				http.Error(w, "Erro ao gerar QR code: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "image/png")
			w.Write(png)
			return
		} else {
			fmt.Println("Evento Login:", evt.Event)
		}
	}
}
