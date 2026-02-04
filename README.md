# 🎮 PromoGames API

API REST desenvolvida em Go para gerenciamento e distribuição automatizada de promoções de jogos através de múltiplas plataformas (Telegram e WhatsApp).

## 📋 Sobre o Projeto

A PromoGames API é um sistema backend robusto que centraliza o gerenciamento de ofertas de jogos e distribui automaticamente para canais do Telegram e WhatsApp. Inclui sistema de wishlist inteligente que notifica usuários quando jogos de seu interesse entram em promoção.

### ✨ Principais Funcionalidades

- 🤖 **Integração com Telegram Bot** - Envio automatizado de promoções e sistema de alertas
- 📱 **Integração com WhatsApp** (whatsmeow) - Distribuição via canais/newsletters
- 🔔 **Sistema de Wishlist** - Usuários recebem notificações de jogos favoritos
- 🖼️ **Gerador de Preview** - Busca automática de imagens de jogos
- 🔒 **Autenticação JWT** - Endpoints protegidos com Supabase Auth
- ⚡ **Rate Limiting** - Proteção contra abuso (100 req/min por IP)
- 🛡️ **Proteções de Segurança** - SSRF, DoS, XSS, Race Conditions

## 🛠️ Tecnologias Utilizadas

- **Go 1.24** - Linguagem principal
- **Supabase** - Banco de dados PostgreSQL e autenticação
- **whatsmeow** - Cliente WhatsApp Web
- **Telegram Bot API** - Integração com Telegram
- **JWT** - Autenticação e autorização
- **CORS** - Suporte para requisições cross-origin

## 📦 Dependências Principais

```go
github.com/golang-jwt/jwt/v5       // Autenticação JWT
github.com/nedpals/supabase-go     // Cliente Supabase
github.com/rs/cors                 // CORS middleware
go.mau.fi/whatsmeow               // WhatsApp client
github.com/jackc/pgx/v5           // Driver PostgreSQL
github.com/skip2/go-qrcode        // Geração de QR codes
```

## 🚀 Instalação e Configuração

### Pré-requisitos

- Go 1.24+
- Conta Supabase
- Bot do Telegram criado via [@BotFather](https://t.me/botfather)
- PostgreSQL (via Supabase)

### Instalação

```bash
# Clone o repositório
git clone https://github.com/GabrielPurificate/PromoGamesAPI.git
cd PromoGamesAPI

# Instale as dependências
go mod download

# Configure as variáveis de ambiente (veja abaixo)
cp .env.example .env

# Execute a aplicação
go run main.go
```

## 🔐 Variáveis de Ambiente

Crie um arquivo `.env` na raiz do projeto:

```env
# Supabase
SUPABASE_URL=https://seu-projeto.supabase.co
SUPABASE_KEY=sua_anon_key
SUPABASE_JWT_SECRET=seu_jwt_secret
SUPABASE_URL_CONNECTION_STRING=postgresql://postgres.[projeto]:[senha]@aws-0-sa-east-1.pooler.supabase.com:5432/postgres

# Telegram
TELEGRAM_BOT_TOKEN=123456789:ABCdefGHIjklMNOpqrsTUVwxyz
TELEGRAM_CHAT_ID=-1001234567890

# Servidor
PORT=8080
```

### ⚠️ Importante: Conexão com Supabase

Para usar prepared statements (melhor performance), use a porta **5432** (conexão direta):
```
postgresql://postgres.[projeto]:[senha]@...supabase.com:5432/postgres
```

A porta 6543 (pooler transaction mode) não suporta prepared statements.

## 📡 Endpoints da API

### Públicos

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| `GET` | `/` | Health check |
| `GET` | `/ping` | Pong response |
| `POST` | `/webhook/telegram` | Webhook para bot do Telegram |

### Protegidos (requer JWT)

| Método | Endpoint | Descrição | Auth |
|--------|----------|-----------|------|
| `GET` | `/check-session` | Valida sessão JWT | ✅ JWT |
| `GET` | `/whatsapp/qr` | Gera QR code para WhatsApp | ✅ JWT |
| `POST` | `/gerar-preview` | Gera preview de promoção | ✅ JWT |
| `POST` | `/enviar-telegram` | Envia mensagem no Telegram | ✅ JWT |

### Exemplos de Uso

#### Gerar Preview de Promoção

```bash
curl -X POST https://sua-api.com/gerar-preview \
  -H "Authorization: Bearer SEU_TOKEN_JWT" \
  -H "Content-Type: application/json" \
  -d '{
    "nome": "God of War Ragnarök",
    "valor": "199.90",
    "loja": "Steam",
    "link": "https://store.steampowered.com/...",
    "cupom": "PROMO10",
    "parcelas": 3,
    "valor_parcela": "66.63",
    "tem_juros": false,
    "is_pix": true
  }'
```

#### Enviar para Telegram

```bash
curl -X POST https://sua-api.com/enviar-telegram \
  -H "Authorization: Bearer SEU_TOKEN_JWT" \
  -H "Content-Type: application/json" \
  -d '{
    "texto": "<b>God of War</b>\n\n💰 R$ 199,90\n\n🔗 Link: https://...",
    "imagem": "https://cdn.exemplo.com/imagem.jpg"
  }'
```

## 🤖 Comandos do Bot Telegram

Os usuários podem interagir com o bot através dos seguintes comandos:

- `/start` - Instruções de uso
- `/desejo [Nome do Jogo]` - Criar alerta de preço
- `/lista` - Ver alertas ativos
- `/remover [Nome do Jogo]` - Remover alerta

## 🔒 Segurança

### Proteções Implementadas

✅ **Autenticação JWT** - Todos endpoints sensíveis protegidos  
✅ **Rate Limiting** - 100 requisições/minuto por IP  
✅ **SSRF Protection** - Validação de URLs, bloqueio de IPs privados  
✅ **DoS Protection** - Limite de tamanho de body (1-2MB)  
✅ **XSS Protection** - Sanitização de inputs  
✅ **Race Condition** - Mutex em operações concorrentes  
✅ **Panic Recovery** - Todas goroutines com defer recover  
✅ **Context Timeout** - Todas operações de rede com timeout

### Validações de Input

- Tamanho de campos (Nome ≤200, Link ≤500, Texto ≤4096)
- URLs apenas HTTPS para imagens
- Limite de 5MB para upload de imagens
- Sanitização de comandos Telegram (3-100 caracteres)

## 🏗️ Estrutura do Projeto

```
PromoGamesAPI/
├── main.go                 # Ponto de entrada, rotas
├── go.mod                  # Dependências
├── api/
│   ├── handlers.go         # Handlers principais
│   ├── handlers_whatsapp.go # Handlers WhatsApp
│   ├── whatsapp.go         # Lógica WhatsApp
│   ├── telegram.go         # Integração Telegram
│   └── wishlist.go         # Sistema de alertas
├── auth/
│   └── middleware.go       # JWT e rate limiting
└── models/
    └── promo.go           # Structs de dados
```

## 📊 Banco de Dados (Supabase)

### Tabelas Necessárias

#### `Jogos`
```sql
CREATE TABLE Jogos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome TEXT NOT NULL,
    slug TEXT UNIQUE NOT NULL,
    image_url TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);
```

#### `alertas_usuario`
```sql
CREATE TABLE alertas_usuario (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    telegram_id BIGINT NOT NULL,
    termo_busca TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```

### Functions (RPC)

```sql
-- Buscar interessados em um produto
CREATE OR REPLACE FUNCTION buscar_interessados(titulo_produto TEXT)
RETURNS TABLE (telegram_id BIGINT, termo_busca TEXT) AS $$
BEGIN
    RETURN QUERY
    SELECT au.telegram_id, au.termo_busca
    FROM alertas_usuario au
    WHERE titulo_produto ILIKE '%' || au.termo_busca || '%';
END;
$$ LANGUAGE plpgsql;

-- Remover wishlist
CREATE OR REPLACE FUNCTION remover_wishlist(p_telegram_id BIGINT, p_termo TEXT)
RETURNS BOOLEAN AS $$
DECLARE
    deleted_count INT;
BEGIN
    DELETE FROM alertas_usuario
    WHERE telegram_id = p_telegram_id
    AND termo_busca ILIKE p_termo;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count > 0;
END;
$$ LANGUAGE plpgsql;
```

## 🚢 Deploy

### Render.com (Recomendado)

1. Conecte seu repositório GitHub
2. Configure as variáveis de ambiente
3. Comando de build: `go build -o bin/main main.go`
4. Comando de start: `./bin/main`

### Docker

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
```

## 📈 Monitoramento

A API expõe os seguintes endpoints para monitoramento:

- `GET /` - Status da API
- `GET /ping` - Health check

## 🤝 Contribuindo

Contribuições são bem-vindas! Por favor:

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/NovaFuncionalidade`)
3. Commit suas mudanças (`git commit -m 'Add: Nova funcionalidade'`)
4. Push para a branch (`git push origin feature/NovaFuncionalidade`)
5. Abra um Pull Request

## 📝 Licença

Este projeto está sob a licença MIT. Veja o arquivo `LICENSE` para mais detalhes.

## 👨‍💻 Autor

**Gabriel Purificate**

- GitHub: [@GabrielPurificate](https://github.com/GabrielPurificate)
- Website: [promogamesbr.com](https://promogamesbr.com)

## 🙏 Agradecimentos

- [whatsmeow](https://github.com/tulir/whatsmeow) - Biblioteca WhatsApp
- [Supabase](https://supabase.com) - Backend as a Service
- Comunidade Go

---

⭐ **Se este projeto foi útil, considere dar uma estrela!**