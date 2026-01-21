package http

import (
	"net/http"

	"socialmeli/internal/domain"
	"socialmeli/internal/service"

	"github.com/gin-gonic/gin"
)

type usersCatalogService interface {
	ListUsers(order string) ([]domain.User, error)
	CreateUser(name string, isSeller bool) (domain.User, error)
}

type UsersCatalogHandlers struct{ us usersCatalogService }

func NewUsersCatalogHandlers(us *service.UserService) *UsersCatalogHandlers {
	return &UsersCatalogHandlers{us: us}
}

type createUserBody struct {
	UserName string `json:"user_name"`
	IsSeller bool   `json:"is_seller"`
}

func (h *UsersCatalogHandlers) List(c *gin.Context) {
	order := c.DefaultQuery("order", domain.NameAsc)
	page, limit, ok := parsePageLimit(c, 20, 100)
	if !ok {
		return
	}
	users, err := h.us.ListUsers(order)
	if err != nil {
		badRequest(c, err)
		return
	}
	usersPage, meta := paginateSlice(users, page, limit)
	c.JSON(http.StatusOK, gin.H{"users": usersPage, "meta": meta})
}

func (h *UsersCatalogHandlers) Create(c *gin.Context) {
	var b createUserBody
	if err := c.ShouldBindJSON(&b); err != nil {
		badRequest(c, err)
		return
	}
	u, err := h.us.CreateUser(b.UserName, b.IsSeller)
	if err != nil {
		badRequest(c, err)
		return
	}
	c.JSON(http.StatusCreated, u)
}
