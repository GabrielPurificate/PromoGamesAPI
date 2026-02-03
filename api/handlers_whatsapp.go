package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/skip2/go-qrcode"
)

func HandlerGetQR(w http.ResponseWriter, r *http.Request) {
	if WAClient == nil {
		http.Error(w, "Cliente WhatsApp não iniciado", 500)
		return
	}

	if WAClient.Store.ID != nil {
		w.Write([]byte("✅ WhatsApp já está conectado! Não precisa de QR."))
		return
	}

	qrChan, _ := WAClient.GetQRChannel(context.Background())
	err := WAClient.Connect()
	if err != nil {
		http.Error(w, "Erro ao conectar: "+err.Error(), 500)
		return
	}

	for evt := range qrChan {
		if evt.Event == "code" {
			png, _ := qrcode.Encode(evt.Code, qrcode.Medium, 256)
			w.Header().Set("Content-Type", "image/png")
			w.Write(png)
			return
		} else {
			fmt.Println("Evento Login:", evt.Event)
		}
	}
}
