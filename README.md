# SocialMeli (Go) - API REST

Implementação completa do desafio SocialMeli em Go, usando Gin + storage em memória.

## Rodar
```bash
go mod tidy
go run ./cmd/api
```

API sobe em `http://localhost:8080`

## Testar
```bash
go test ./... -v
```

## Endpoints (User Stories)

- US-0001 Follow
  - POST `/users/{userId}/follow/{userIdToFollow}`

- US-0002 Followers count
  - GET `/users/{userId}/followers/count`

- US-0003 Followers list (+ order name_asc/name_desc)
  - GET `/users/{userId}/followers/list?order=name_asc`

- US-0004 Followed list (+ order name_asc/name_desc)
  - GET `/users/{userId}/followed/list?order=name_desc`

- US-0005 Publish product
  - POST `/products/publish`

- US-0006 Posts from followed sellers in last 2 weeks (+ order date_asc/date_desc)
  - GET `/products/followed/{userId}/list?order=date_desc`

- US-0007 Unfollow
  - POST `/users/{userId}/unfollow/{userIdToUnfollow}`

- US-0010 Promo publish
  - POST `/products/promo-pub`

- US-0011 Promo count
  - GET `/products/promo-pub/count?user_id={userId}`

- US-0012 Promo list
  - GET `/products/promo-pub/list?user_id={userId}`

## Seed (para testar rápido)
O servidor já sobe com alguns usuários em memória:
- 123 (usuario123)
- 234 (vendedor1)
- 6932 (vendedor2)
- 4698 (usuario1)

Você pode alterar em `internal/store/seed.go`.
