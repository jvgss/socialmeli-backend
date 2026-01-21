package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"socialmeli/internal/service"

	"github.com/gin-gonic/gin"
)

/*
	Teste de rotas
	- garante que o router sobe
	- garante que TODAS as rotas existem
	- não testa regra de negócio
*/

func TestNewRouter_AllRoutesExist(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// services "vazios" (zero-value)
	us := &service.UserService{}
	ps := &service.ProductService{}
	as := &service.AuthService{}

	router := NewRouter(us, ps, as)
	if router == nil {
		t.Fatalf("router should not be nil")
	}

	tests := []struct {
		method string
		path   string
	}{
		// USERS
		{http.MethodPost, "/users/abc/follow/1"},       // param inválido
		{http.MethodPost, "/users/1/unfollow/abc"},     // param inválido
		{http.MethodGet, "/users/abc/followers/count"}, // param inválido
		{http.MethodGet, "/users/abc/followers/list"},  // param inválido
		{http.MethodGet, "/users/abc/followed/list"},   // param inválido

		// PRODUCTS
		{http.MethodPost, "/products/publish"},          // body vazio → 400
		{http.MethodGet, "/products/followed/abc/list"}, // param inválido
		{http.MethodPost, "/products/promo-pub"},        // body vazio → 400
		{http.MethodGet, "/products/promo-pub/count?user_id=abc"},
		{http.MethodGet, "/products/promo-pub/list?user_id=abc"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// O importante NÃO é o status exato,
		// mas que NÃO seja 404 (rota inexistente)
		if w.Code == http.StatusNotFound {
			t.Fatalf("route not found: %s %s", tt.method, tt.path)
		}
	}
}
