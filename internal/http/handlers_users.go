package http

import (
	"net/http"
	"strconv"

	"socialmeli/internal/domain"
	"socialmeli/internal/service"

	"github.com/gin-gonic/gin"
)

// interface mínima para permitir mock em testes
type userService interface {
	Follow(userID, sellerID int) error
	Unfollow(userID, sellerID int) error
	FollowersCount(sellerID int) (domain.User, int, error)
	FollowersList(sellerID int, order string) (domain.User, []domain.User, error)
	FollowedList(userID int, order string) (domain.User, []domain.User, error)
}

type UserHandlers struct{ us userService }

// mantém compatível com o resto do projeto: você continua passando *service.UserService
func NewUserHandlers(us *service.UserService) *UserHandlers { return &UserHandlers{us: us} }

// (opcional, mas útil para testes) permite injetar mock diretamente sem precisar *service.UserService
func NewUserHandlersWithService(us userService) *UserHandlers { return &UserHandlers{us: us} }

func mustIntParam(c *gin.Context, name string) (int, bool) {
	v := c.Param(name)
	i, err := strconv.Atoi(v)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parâmetro inválido: " + name})
		return 0, false
	}
	return i, true
}

// Follow godoc
// @Summary Seguir um usuário
// @Description Um usuário (userId) passa a seguir outro usuário (userIdToFollow)
// @Tags users
// @Produce json
// @Param userId path int true "ID do usuário que vai seguir"
// @Param userIdToFollow path int true "ID do usuário a ser seguido"
// @Success 200 {object} map[string]string "OK"
// @Failure 400 {object} map[string]string "Parâmetros inválidos ou regra de negócio"
// @Router /users/{userId}/follow/{userIdToFollow} [post]
func (h *UserHandlers) Follow(c *gin.Context) {
	userID, ok := mustIntParam(c, "userId")
	if !ok {
		return
	}
	sellerID, ok := mustIntParam(c, "userIdToFollow")
	if !ok {
		return
	}

	if err := h.us.Follow(userID, sellerID); err != nil {
		badRequest(c, err)
		return
	}
	okNoBody(c)
}

// Unfollow godoc
// @Summary Deixar de seguir um usuário
// @Description Um usuário (userId) deixa de seguir outro usuário (userIdToUnfollow)
// @Tags users
// @Produce json
// @Param userId path int true "ID do usuário que vai deixar de seguir"
// @Param userIdToUnfollow path int true "ID do usuário a deixar de seguir"
// @Success 200 {object} map[string]string "OK"
// @Failure 400 {object} map[string]string "Parâmetros inválidos ou regra de negócio"
// @Router /users/{userId}/unfollow/{userIdToUnfollow} [post]
func (h *UserHandlers) Unfollow(c *gin.Context) {
	userID, ok := mustIntParam(c, "userId")
	if !ok {
		return
	}
	sellerID, ok := mustIntParam(c, "userIdToUnfollow")
	if !ok {
		return
	}

	if err := h.us.Unfollow(userID, sellerID); err != nil {
		badRequest(c, err)
		return
	}
	okNoBody(c)
}

// FollowersCount godoc
// @Summary Contar seguidores
// @Description Retorna a quantidade de seguidores de um usuário (seller)
// @Tags users
// @Produce json
// @Param userId path int true "ID do usuário (seller)"
// @Success 200 {object} FollowersCountResponse
// @Failure 400 {object} map[string]string "Parâmetro inválido ou usuário não encontrado"
// @Router /users/{userId}/followers/count [get]
func (h *UserHandlers) FollowersCount(c *gin.Context) {
	sellerID, ok := mustIntParam(c, "userId")
	if !ok {
		return
	}

	u, count, err := h.us.FollowersCount(sellerID)
	if err != nil {
		badRequest(c, err)
		return
	}

	c.JSON(http.StatusOK, FollowersCountResponse{
		UserID: u.ID, UserName: u.Name, FollowersCount: count,
	})
}

// FollowersList godoc
// @Summary Listar seguidores
// @Description Retorna a lista de seguidores de um usuário. Pode ordenar por nome.
// @Tags users
// @Produce json
// @Param userId path int true "ID do usuário (seller)"
// @Param order query string false "Ordenação por nome" Enums(asc,desc)
// @Success 200 {object} FollowersListResponse
// @Failure 400 {object} map[string]string "Parâmetro inválido, order inválida ou usuário não encontrado"
// @Router /users/{userId}/followers/list [get]
func (h *UserHandlers) FollowersList(c *gin.Context) {
	sellerID, ok := mustIntParam(c, "userId")
	if !ok {
		return
	}
	order := c.Query("order")

	u, followers, err := h.us.FollowersList(sellerID, order)
	if err != nil {
		badRequest(c, err)
		return
	}

	respFollowers := make([]SimpleUser, 0, len(followers))
	for _, f := range followers {
		respFollowers = append(respFollowers, SimpleUser{UserID: f.ID, UserName: f.Name})
	}

	c.JSON(http.StatusOK, FollowersListResponse{
		UserID: u.ID, UserName: u.Name, Followers: respFollowers,
	})
}

// FollowedList godoc
// @Summary Listar seguidos
// @Description Retorna a lista de usuários que um userId segue. Pode ordenar por nome.
// @Tags users
// @Produce json
// @Param userId path int true "ID do usuário"
// @Param order query string false "Ordenação por nome" Enums(asc,desc)
// @Success 200 {object} FollowedListResponse
// @Failure 400 {object} map[string]string "Parâmetro inválido, order inválida ou usuário não encontrado"
// @Router /users/{userId}/followed/list [get]
func (h *UserHandlers) FollowedList(c *gin.Context) {
	userID, ok := mustIntParam(c, "userId")
	if !ok {
		return
	}
	order := c.Query("order")

	u, followed, err := h.us.FollowedList(userID, order)
	if err != nil {
		badRequest(c, err)
		return
	}

	resp := make([]SimpleUser, 0, len(followed))
	for _, f := range followed {
		resp = append(resp, SimpleUser{UserID2: f.ID, UserName2: f.Name})
	}

	c.JSON(http.StatusOK, FollowedListResponse{
		UserID: u.ID, UserName: u.Name, Followed: resp,
	})
}
