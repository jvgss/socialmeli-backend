package http

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"socialmeli/internal/domain"
	"socialmeli/internal/service"
	"socialmeli/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestProfileHandlers_MeAndMyPosts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	st := store.NewMemoryStore()
	// cria conta (tambem cria user social)
	acc, err := st.CreateAccount("User", "user@example.com", "hash", false)
	require.NoError(t, err)

	// posts
	_, err = st.AddPost(domain.Post{UserID: acc.ID, Product: domain.Product{ProductID: 1}, Category: 1, Price: 10, HasPromo: false, Date: time.Now().Add(-24 * time.Hour)})
	require.NoError(t, err)
	_, err = st.AddPost(domain.Post{UserID: acc.ID, Product: domain.Product{ProductID: 2}, Category: 2, Price: 20, HasPromo: true, Date: time.Now()})
	require.NoError(t, err)

	us := service.NewUserService(st)
	ph := NewProfileHandlers(us)

	r := gin.New()
	// "auth" fake
	r.GET("/me", func(c *gin.Context) { c.Set("auth_user_id", acc.ID); ph.Me(c) })
	r.GET("/myposts", func(c *gin.Context) { c.Set("auth_user_id", acc.ID); ph.MyPosts(c) })
	r.GET("/me-noauth", ph.Me)

	// sem auth
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/me-noauth", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)

	// me ok
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/me", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "\"followers_count\"")
	require.Contains(t, w.Body.String(), "\"email\":\"user@example.com\"")

	// posts default order=date_desc, entao o post mais recente deve aparecer primeiro
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/myposts", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	// product_id 2 (mais recente) deve vir antes
	body := w.Body.String()
	require.Less(t, strings.Index(body, "\"product_id\":2"), strings.Index(body, "\"product_id\":1"))
}

func TestProfileHandlers_UploadAvatar_ValidationsAndSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	st := store.NewMemoryStore()
	acc, err := st.CreateAccount("User", "user@example.com", "hash", false)
	require.NoError(t, err)

	us := service.NewUserService(st)
	ph := NewProfileHandlers(us)

	// isola filesystem em um diretório temporário
	tmp := t.TempDir()
	oldWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(tmp))
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	r := gin.New()
	r.POST("/avatar", func(c *gin.Context) { c.Set("auth_user_id", acc.ID); ph.UploadAvatar(c) })

	// 1) arquivo grande > 2MB
	{
		var buf bytes.Buffer
		mw := multipart.NewWriter(&buf)
		fw, err := mw.CreateFormFile("avatar", "big.jpg")
		require.NoError(t, err)
		fw.Write(bytes.Repeat([]byte("a"), (2<<20)+1))
		mw.Close()

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/avatar", &buf)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusBadRequest, w.Code)
	}

	// 2) extensão inválida
	{
		var buf bytes.Buffer
		mw := multipart.NewWriter(&buf)
		fw, err := mw.CreateFormFile("avatar", "x.exe")
		require.NoError(t, err)
		fw.Write([]byte("abc"))
		mw.Close()

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/avatar", &buf)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusBadRequest, w.Code)
	}

	// 3) sucesso (png)
	{
		var buf bytes.Buffer
		mw := multipart.NewWriter(&buf)
		fw, err := mw.CreateFormFile("avatar", "avatar.png")
		require.NoError(t, err)
		fw.Write([]byte{0x89, 0x50, 0x4e, 0x47})
		mw.Close()

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/avatar", &buf)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), "/static/avatars/")

		// arquivo salvo
		saved := filepath.Join("uploads", "avatars", fmt.Sprintf("%d.png", acc.ID))
		_, err = os.Stat(saved)
		require.NoError(t, err)

		// avatar atualizado na conta
		updated, ok := st.GetAccount(acc.ID)
		require.True(t, ok)
		require.Contains(t, updated.AvatarURL, "/static/avatars/")
	}
}

func TestProfileHandlers_GetProfile_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	st := store.NewMemoryStore()
	acc, err := st.CreateAccount("User", "user@example.com", "hash", false)
	require.NoError(t, err)

	// adiciona alguns seguidores e seguidos
	st.SeedUsers([]domain.User{
		{ID: 2, Name: "Follower1", IsSeller: false},
		{ID: 3, Name: "Follower2", IsSeller: false},
		{ID: 4, Name: "Followed1", IsSeller: true},
	})
	_ = st.Follow(2, acc.ID)
	_ = st.Follow(3, acc.ID)
	_ = st.Follow(acc.ID, 4)

	us := service.NewUserService(st)
	ph := NewProfileHandlers(us)

	r := gin.New()
	r.GET("/users/:userId/profile", ph.GetProfile)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users/"+strconv.Itoa(acc.ID)+"/profile", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "\"followers_count\"")
	require.Contains(t, w.Body.String(), "\"followed_count\"")
	require.Contains(t, w.Body.String(), "\"publications_count\"")
}

func TestProfileHandlers_GetProfile_UserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	st := store.NewMemoryStore()
	us := service.NewUserService(st)
	ph := NewProfileHandlers(us)

	r := gin.New()
	r.GET("/users/:userId/profile", ph.GetProfile)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users/999/profile", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}
