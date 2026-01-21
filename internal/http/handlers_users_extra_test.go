package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"socialmeli/internal/domain"
	"socialmeli/internal/service"
	"socialmeli/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestUsersCatalogHandlers_List_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	st := store.NewMemoryStore()
	st.SeedUsers([]domain.User{
		{ID: 1, Name: "Alice", IsSeller: false},
		{ID: 2, Name: "Bob", IsSeller: true},
	})
	us := service.NewUserService(st)
	h := NewUsersCatalogHandlers(us)

	r := gin.New()
	r.GET("/users", h.List)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users?order=name_asc&page=1&limit=10", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "\"users\"")
	require.Contains(t, w.Body.String(), "\"meta\"")
}

func TestUsersCatalogHandlers_List_InvalidOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	st := store.NewMemoryStore()
	us := service.NewUserService(st)
	h := NewUsersCatalogHandlers(us)

	r := gin.New()
	r.GET("/users", h.List)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users?order=invalid", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUsersCatalogHandlers_Create_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	st := store.NewMemoryStore()
	us := service.NewUserService(st)
	h := NewUsersCatalogHandlers(us)

	r := gin.New()
	r.POST("/users", h.Create)

	body, _ := json.Marshal(map[string]any{
		"user_name": "New User",
		"is_seller": true,
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	require.Contains(t, w.Body.String(), "\"user_name\":\"New User\"")
}

func TestUsersCatalogHandlers_Create_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	st := store.NewMemoryStore()
	us := service.NewUserService(st)
	h := NewUsersCatalogHandlers(us)

	r := gin.New()
	r.POST("/users", h.Create)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUsersCatalogHandlers_Create_InvalidName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	st := store.NewMemoryStore()
	us := service.NewUserService(st)
	h := NewUsersCatalogHandlers(us)

	r := gin.New()
	r.POST("/users", h.Create)

	body, _ := json.Marshal(map[string]any{
		"user_name": "",
		"is_seller": false,
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}
