# SocialMeli 🟡🔵 — API REST em Go

API REST desenvolvida em **Go** como implementação completa do desafio **SocialMeli**,
simulando funcionalidades essenciais de um marketplace inspirado no **Mercado Livre**.

O foco do projeto é **arquitetura limpa**, **boas práticas**, **testabilidade** e
**clareza de regras de negócio**.

---

## 🚀 Stack & Tecnologias
- **Go**
- **Gin** (HTTP framework)
- **Arquitetura em camadas**
- **Testes unitários com `go test`**
- **Storage em memória** (facilmente substituível por banco relacional)

---

## 🧠 Conceitos aplicados
- Separação clara de responsabilidades (Handler → Service → Store → Domain)
- Validações centralizadas no domínio
- Regras de negócio isoladas da camada HTTP
- Cobertura de testes focada em regras críticas
- Código preparado para evolução (ex: troca de storage)

---

## 🏗️ Arquitetura
cmd/api
└── main.go # bootstrap da aplicação

internal/
├── http/ # handlers HTTP (controllers)
├── service/ # regras de negócio
├── store/ # acesso a dados (memory store)
└── domain/ # entidades, validações e ordenações

shell
Copiar código

### Fluxo padrão
Request HTTP
→ Handler
→ Service
→ Store
→ Resposta

yaml
Copiar código

---

## 📦 Funcionalidades implementadas

### Usuários
- Seguir e deixar de seguir usuários
- Contagem de seguidores
- Listagem de seguidores e seguindo
- Ordenação por nome (asc / desc)

### Produtos
- Publicação de produtos
- Feed de produtos de vendedores seguidos (últimas 2 semanas)
- Publicação de promoções
- Cálculo de preço final com desconto
- Listagem e contagem de promoções

---

## 🔌 Endpoints (User Stories)

### US-0001 — Follow
POST /users/{userId}/follow/{userIdToFollow}

shell
Copiar código

### US-0002 — Followers count
GET /users/{userId}/followers/count

shell
Copiar código

### US-0003 — Followers list
GET /users/{userId}/followers/list?order=name_asc

shell
Copiar código

### US-0004 — Followed list
GET /users/{userId}/followed/list?order=name_desc

shell
Copiar código

### US-0005 — Publish product
POST /products/publish

shell
Copiar código

### US-0006 — Feed (últimas 2 semanas)
GET /products/followed/{userId}/list?order=date_desc

shell
Copiar código

### US-0007 — Unfollow
POST /users/{userId}/unfollow/{userIdToUnfollow}

shell
Copiar código

### US-0010 — Promo publish
POST /products/promo-pub

shell
Copiar código

### US-0011 — Promo count
GET /products/promo-pub/count?user_id={userId}

shell
Copiar código

### US-0012 — Promo list
GET /products/promo-pub/list?user_id={userId}

yaml
Copiar código

---

## ▶️ Como rodar o projeto

### Subir a API
```bash
# Como rodar
```bash
go mod tidy
DATABASE_URL="postgres://socialmeli:socialmeli@localhost:5432/socialmeli?sslmode=disable" \
go run ./cmd/api


API disponível em: http://localhost:8080


🧪 Testes
bash
go test ./... -v


O projeto possui testes unitários focados em:

regras de negócio

validações

ordenação

cenários de erro

🌱 Seed de dados
Ao iniciar a aplicação, alguns usuários são criados automaticamente:

123 — usuario123

234 — vendedor1

6932 — vendedor2

4698 — usuario1

O seed pode ser alterado em:

swift
Copiar código
internal/store/seed.go
🔮 Próximos passos (roadmap)
Persistência em banco relacional (PostgreSQL)

Autenticação (JWT)

Paginação e filtros

Upload real de imagens

Observabilidade (logs estruturados)

👨‍💻 Autor
Projeto desenvolvido para estudo, prática de arquitetura backend
