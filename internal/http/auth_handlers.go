package http

import (
	"net/http"
	"strings"
	"time"

	"socialmeli/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandlers struct {
	as *service.AuthService
}

func NewAuthHandlers(as *service.AuthService) *AuthHandlers { return &AuthHandlers{as: as} }

type tokenResponse struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
}

func (h *AuthHandlers) Register(c *gin.Context) {
	var p service.RegisterPayload
	if err := c.ShouldBindJSON(&p); err != nil {
		badRequest(c, err)
		return
	}
	acc, err := h.as.Register(p)
	if err != nil {
		badRequest(c, err)
		return
	}
	token, err := MakeToken(acc.ID, 2*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao gerar token"})
		return
	}
	c.JSON(http.StatusCreated, tokenResponse{Token: token, User: acc})
}

func (h *AuthHandlers) Login(c *gin.Context) {
	var p service.LoginPayload
	if err := c.ShouldBindJSON(&p); err != nil {
		badRequest(c, err)
		return
	}
	acc, err := h.as.Login(p)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	token, err := MakeToken(acc.ID, 2*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao gerar token"})
		return
	}
	c.JSON(http.StatusOK, tokenResponse{Token: token, User: acc})
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token ausente"})
			return
		}
		tokenStr := strings.TrimSpace(auth[7:])
		claims, err := ParseToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.Set("auth_user_id", claims.Sub)
		c.Next()
	}
}
