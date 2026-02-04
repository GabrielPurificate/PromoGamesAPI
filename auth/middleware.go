package auth

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	rateLimitMap   = make(map[string][]time.Time)
	rateLimitMutex sync.Mutex
)

func MiddlewareJWT(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			next(w, r)
			return
		}

		jwtSecret := os.Getenv("SUPABASE_JWT_SECRET")
		if jwtSecret == "" {
			fmt.Println("CRITICAL: SUPABASE_JWT_SECRET não configurado no ambiente")
			http.Error(w, "Erro interno de configuração", http.StatusInternalServerError)
			return
		}

		clientIP := getClientIP(r)
		if !checkRateLimit(clientIP, 100, time.Minute) {
			http.Error(w, "Rate limit excedido", http.StatusTooManyRequests)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("método de assinatura inesperado: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			fmt.Printf("Erro de Token: %v\n", err)
			http.Error(w, "Token inválido ou expirado", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func getClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return strings.Split(r.RemoteAddr, ":")[0]
}

func checkRateLimit(key string, maxRequests int, window time.Duration) bool {
	rateLimitMutex.Lock()
	defer rateLimitMutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-window)

	if times, exists := rateLimitMap[key]; exists {
		var validTimes []time.Time
		for _, t := range times {
			if t.After(cutoff) {
				validTimes = append(validTimes, t)
			}
		}
		rateLimitMap[key] = validTimes
	}

	if len(rateLimitMap[key]) >= maxRequests {
		return false
	}

	rateLimitMap[key] = append(rateLimitMap[key], now)
	return true
}
