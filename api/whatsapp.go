package api

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

var (
	WAClient      *whatsmeow.Client
	waClientMutex sync.RWMutex
)

func ConectarWhatsApp() error {
	dbURL := os.Getenv("SUPABASE_URL_CONNECTION_STRING")

	dbLog := waLog.Stdout("Database", "ERROR", true)

	if dbURL != "" {
		if dbURL[len(dbURL)-1] != '?' && !contains(dbURL, "?") {
			dbURL += "?"
		} else if dbURL[len(dbURL)-1] != '&' {
			dbURL += "&"
		}
		dbURL += "sslmode=require"
	}

	container, err := sqlstore.New(context.Background(), "pgx", dbURL, dbLog)
	if err != nil {
		return fmt.Errorf("erro ao conectar no banco para whatsapp: %v", err)
	}

	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil {
		return fmt.Errorf("erro ao pegar device store: %v", err)
	}

	clientLog := waLog.Stdout("Client", "INFO", true)

	waClientMutex.Lock()
	WAClient = whatsmeow.NewClient(deviceStore, clientLog)
	waClientMutex.Unlock()

	if WAClient.Store.ID == nil {
		fmt.Println("WHATSAPP: Nenhum login encontrado. Acesse /whatsapp/qr para logar.")
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err = WAClient.Connect()
		if err != nil {
			return fmt.Errorf("erro ao reconectar whatsapp: %v", err)
		}
		_ = ctx
		fmt.Println("WHATSAPP: Conectado.")
	}

	return nil
}

func EnviarMensagemCanal(newsletterID string, texto string, imagemURL string) error {
	waClientMutex.RLock()
	defer waClientMutex.RUnlock()

	if WAClient == nil || !WAClient.IsConnected() {
		return fmt.Errorf("whatsapp desconectado")
	}

	jid, err := types.ParseJID(newsletterID)
	if err != nil {
		return fmt.Errorf("newsletter ID inválido: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if imagemURL == "" {
		_, err := WAClient.SendMessage(ctx, jid, &waE2E.Message{
			Conversation: proto.String(texto),
		})
		return err
	}

	if err := validarURLSegura(imagemURL); err != nil {
		return err
	}

	resp, err := http.Get(imagemURL)
	if err != nil {
		return fmt.Errorf("erro ao baixar imagem: %v", err)
	}
	defer resp.Body.Close()

	imgData, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		return fmt.Errorf("erro ao ler imagem: %v", err)
	}
	if len(imgData) >= 5*1024*1024 {
		return fmt.Errorf("imagem muito grande (máx 5MB)")
	}

	mimeType := http.DetectContentType(imgData)
	fmt.Printf("WHATSAPP: Enviando imagem com MIME type detectado: %s\n", mimeType)

	uploadResp, err := WAClient.Upload(ctx, imgData, whatsmeow.MediaImage)
	if err != nil {
		return fmt.Errorf("erro upload whatsapp: %v", err)
	}

	msg := &waE2E.Message{
		ImageMessage: &waE2E.ImageMessage{
			Caption:       proto.String(texto),
			URL:           proto.String(uploadResp.URL),
			DirectPath:    proto.String(uploadResp.DirectPath),
			MediaKey:      uploadResp.MediaKey,
			Mimetype:      proto.String(mimeType),
			FileEncSHA256: uploadResp.FileEncSHA256,
			FileSHA256:    uploadResp.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(imgData))),
		},
	}

	_, err = WAClient.SendMessage(ctx, jid, msg)
	return err
}

func ListarCanais() {
	waClientMutex.RLock()
	defer waClientMutex.RUnlock()

	if WAClient == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	newsletters, err := WAClient.GetSubscribedNewsletters(ctx)
	if err != nil {
		fmt.Println("Erro ao listar canais:", err)
		return
	}

	fmt.Println("--- MEUS CANAIS ---")
	for _, news := range newsletters {
		fmt.Printf("NEWSLETTER INFO: %+v\n", news)
	}
	fmt.Println("-----------------------")
}

func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func validarURLSegura(urlStr string) error {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("URL inválida: %v", err)
	}

	if parsedURL.Scheme != "https" {
		return fmt.Errorf("apenas URLs HTTPS são permitidas")
	}

	hostname := parsedURL.Hostname()

	ips, err := net.LookupIP(hostname)
	if err != nil {
		return fmt.Errorf("erro ao resolver DNS: %v", err)
	}

	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsPrivate() {
			return fmt.Errorf("acesso a rede privada bloqueado: %s", ip.String())
		}
	}

	return nil
}
