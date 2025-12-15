package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/nedpals/supabase-go"
	"github.com/rs/cors"

	"github.com/GabrielPurificate/PromoGamesAPI/api"
)

func main() {

	_ = godotenv.Load()

	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	if supabaseUrl == "" || supabaseKey == "" {
		log.Fatal("SUPABASE_URL and SUPABASE_KEY environment variables must be set")
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

	mux.HandleFunc("/gerar-preview", api.HandlerGerarPreview(client))

	mux.HandleFunc("/enviar-telegram", api.HandlerEnviarTelegram)

	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}).Handler(mux)

	fmt.Printf("🚀 Servidor rodando em http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}
