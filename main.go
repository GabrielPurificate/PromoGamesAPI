package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/nedpals/supabase-go"
	"github.com/rs/cors"

	"github.com/GabrielPurificate/PromoGamesAPI/api"
	"github.com/GabrielPurificate/PromoGamesAPI/auth"
)

func main() {

	_ = godotenv.Load()

	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	if os.Getenv("SUPABASE_JWT_SECRET") == "" {
		log.Println("AVISO: SUPABASE_JWT_SECRET não encontrado. O Middleware de Auth vai falhar se for acionado.")
	}

	if supabaseUrl == "" || supabaseKey == "" {
		log.Fatal("SUPABASE_URL and SUPABASE_KEY environment variables must be set")
	}

	if err := api.ConectarWhatsApp(); err != nil {
		log.Println("Erro fatal WhatsApp:", err)
	}

	client := supabase.CreateClient(supabaseUrl, supabaseKey)

	mux := http.NewServeMux()

	mux.HandleFunc("/ping", api.HandlerPing)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "online",
			"msg":    "API PostPanel rodando",
		})
	})

	mux.HandleFunc("/check-session", auth.MiddlewareJWT(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{
			"valid": true,
		})
	}))

	mux.HandleFunc("/whatsapp/qr", auth.MiddlewareJWT(api.HandlerGetQR))

	mux.HandleFunc("/gerar-preview", auth.MiddlewareJWT(api.HandlerGerarPreview(client)))

	mux.HandleFunc("/webhook/telegram", api.HandlerWebhookTelegram(client))

	mux.HandleFunc("/enviar-telegram", auth.MiddlewareJWT(api.HandlerEnviarTelegram(client)))

	mux.HandleFunc("/buscar-jogos", auth.MiddlewareJWT(api.HandlerBuscarJogos(client)))
	mux.HandleFunc("/cupons-recentes", auth.MiddlewareJWT(api.HandlerCuponsRecentes(client)))
	mux.HandleFunc("/jogos-populares", auth.MiddlewareJWT(api.HandlerJogosPopulares(client)))

	//api.ListarCanais()

	allowedOrigins := []string{"*"}
	if envOrigins := os.Getenv("CORS_ALLOWED_ORIGINS"); envOrigins != "" {
		allowedOrigins = strings.Split(envOrigins, ",")
	} else {
		log.Println("AVISO: CORS_ALLOWED_ORIGINS não definido. Permitindo todas as origens (*).")
	}

	handler := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}).Handler(mux)

	fmt.Printf("Servidor rodando em http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}
