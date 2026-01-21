package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"socialmeli/internal/service"
	"socialmeli/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAuthHandlers_RegisterAndLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	st := store.NewMemoryStore()
	as := service.NewAuthService(st)
	ah := NewAuthHandlers(as)

	r := gin.New()
	r.POST("/auth/register", ah.Register)
	r.POST("/auth/login", ah.Login)

	// register
	body, _ := json.Marshal(map[string]any{
		"name":      "User 1",
		"email":     "u1@example.com",
		"password":  "123456",
		"is_seller": true,
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
	require.Contains(t, w.Body.String(), "\"token\"")
	require.Contains(t, w.Body.String(), "\"email\":\"u1@example.com\"")

	// login ok
	body, _ = json.Marshal(map[string]any{"email": "u1@example.com", "password": "123456"})
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "\"token\"")

	// login invalid
	body, _ = json.Marshal(map[string]any{"email": "u1@example.com", "password": "wrong"})
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "test-secret")

	r := gin.New()
	r.GET("/protected", AuthMiddleware(), func(c *gin.Context) {
		v, _ := c.Get("auth_user_id")
		c.JSON(200, gin.H{"uid": v})
	})

	// missing token
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)

	// invalid token
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)

	// ok
	tok, err := MakeToken(77, time.Hour)
	require.NoError(t, err)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "\"uid\":77")
}
