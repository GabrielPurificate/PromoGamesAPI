package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

var WAClient *whatsmeow.Client

func ConectarWhatsApp() error {
	dbURL := os.Getenv("SUPABASE_URL_CONNECTION_STRING")

	dbLog := waLog.Stdout("Database", "ERROR", true)

	container, err := sqlstore.New(context.Background(), "postgres", dbURL, dbLog)
	if err != nil {
		return fmt.Errorf("erro ao conectar no banco para whatsapp: %v", err)
	}

	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil {
		return fmt.Errorf("erro ao pegar device store: %v", err)
	}

	clientLog := waLog.Stdout("Client", "INFO", true)
	WAClient = whatsmeow.NewClient(deviceStore, clientLog)

	if WAClient.Store.ID == nil {
		fmt.Println("⚠️ WHATSAPP: Nenhum login encontrado. Acesse /whatsapp/qr para logar.")
	} else {
		err = WAClient.Connect()
		if err != nil {
			return fmt.Errorf("erro ao reconectar whatsapp: %v", err)
		}
		fmt.Println("✅ WHATSAPP: Conectado e Logado!")
	}

	return nil
}

func EnviarMensagemCanal(newsletterID string, texto string, imagemURL string) error {
	if WAClient == nil || !WAClient.IsConnected() {
		return fmt.Errorf("whatsapp desconectado")
	}

	jid, _ := types.ParseJID(newsletterID)

	if imagemURL == "" {
		_, err := WAClient.SendMessage(context.Background(), jid, &waE2E.Message{
			Conversation: proto.String(texto),
		})
		return err
	}

	resp, err := http.Get(imagemURL)
	if err != nil {
		return fmt.Errorf("erro ao baixar imagem: %v", err)
	}
	defer resp.Body.Close()

	imgData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("erro ao ler imagem: %v", err)
	}

	uploadResp, err := WAClient.Upload(context.Background(), imgData, whatsmeow.MediaImage)
	if err != nil {
		return fmt.Errorf("erro upload whatsapp: %v", err)
	}

	msg := &waE2E.Message{
		ImageMessage: &waE2E.ImageMessage{
			Caption:       proto.String(texto),
			URL:           proto.String(uploadResp.URL),
			DirectPath:    proto.String(uploadResp.DirectPath),
			MediaKey:      uploadResp.MediaKey,
			Mimetype:      proto.String("image/jpeg"),
			FileEncSHA256: uploadResp.FileEncSHA256,
			FileSHA256:    uploadResp.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(imgData))),
		},
	}

	_, err = WAClient.SendMessage(context.Background(), jid, msg)
	return err
}

func ListarCanais() {
	if WAClient == nil {
		return
	}

	newsletters, err := WAClient.GetSubscribedNewsletters(context.Background())
	if err != nil {
		fmt.Println("Erro ao listar canais:", err)
		return
	}

	fmt.Println("📢 --- MEUS CANAIS ---")
	for _, news := range newsletters {
		fmt.Printf("NEWSLETTER INFO: %+v\n", news)
	}
	fmt.Println("-----------------------")
}
