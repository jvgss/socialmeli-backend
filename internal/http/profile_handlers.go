package http

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"socialmeli/internal/service"

	"github.com/gin-gonic/gin"
)

type ProfileHandlers struct {
	us *service.UserService
}

func NewProfileHandlers(us *service.UserService) *ProfileHandlers { return &ProfileHandlers{us: us} }

func (h *ProfileHandlers) Me(c *gin.Context) {
	uid, ok := c.Get("auth_user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token ausente"})
		return
	}
	prof, err := h.us.GetProfile(uid.(int))
	if err != nil {
		badRequest(c, err)
		return
	}
	c.JSON(http.StatusOK, prof)
}

func (h *ProfileHandlers) GetProfile(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("userId"))
	prof, err := h.us.GetProfile(id)
	if err != nil {
		badRequest(c, err)
		return
	}
	c.JSON(http.StatusOK, prof)
}

func (h *ProfileHandlers) MyPosts(c *gin.Context) {
	uid, ok := c.Get("auth_user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token ausente"})
		return
	}
	order := c.DefaultQuery("order", "date_desc")
	posts, err := h.us.PostsByUser(uid.(int), order)
	if err != nil {
		badRequest(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"posts": posts})
}

func (h *ProfileHandlers) UploadAvatar(c *gin.Context) {
	uidAny, ok := c.Get("auth_user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token ausente"})
		return
	}
	uid := uidAny.(int)

	file, err := c.FormFile("avatar")
	if file.Size > 2<<20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "avatar muito grande (max 2MB)"})
		return
	}
	if err != nil {
		badRequest(c, err)
		return
	}

	uploadDir := "uploads/avatars"
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao criar pasta"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == "" {
		ext = ".jpg"
	}
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
		// ok
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "formato inválido (use jpg, png ou webp)"})
		return
	}
	localPath := filepath.Join(uploadDir, strconv.Itoa(uid)+ext)
	if err := c.SaveUploadedFile(file, localPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao salvar"})
		return
	}

	avatarURL := "/static/avatars/" + strconv.Itoa(uid) + ext
	acc, err := h.us.UpdateAvatar(uid, avatarURL)
	if err != nil {
		badRequest(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": acc})
}
