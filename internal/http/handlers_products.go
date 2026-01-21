package http

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"socialmeli/internal/domain"
	"socialmeli/internal/service"

	"github.com/gin-gonic/gin"
)

type ProductService interface {
	Publish(service.PublishPayload) (int, error)
	FollowedLastTwoWeeks(userID int, order string) ([]domain.Post, error)
	PromoCount(userID int) (domain.User, int, error)
	PromoList(userID int) (domain.User, []domain.Post, error)
	DeleteMyPost(userID, postID int) error
}

type ProductHandlers struct {
	ps ProductService
}

func NewProductHandlers(ps ProductService) *ProductHandlers {
	return &ProductHandlers{ps: ps}
}

// Publish godoc
// @Summary Publica um novo produto
// @Description Publica um produto sem promoção (HasPromo=false e Discount=0)
// @Tags products
// @Accept json
// @Produce json
// @Param payload body service.PublishPayload true "Dados do produto"
// @Success 200 {object} PublishResponse
// @Failure 400 {object} map[string]string "JSON inválido"
// @Router /products/publish [post]
func (h *ProductHandlers) Publish(c *gin.Context) {
	var payload service.PublishPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}

	payload.HasPromo = false
	payload.Discount = 0

	postID, err := h.ps.Publish(payload)
	if err != nil {
		badRequest(c, err)
		return
	}

	c.JSON(http.StatusOK, PublishResponse{PostID: postID})
}

// PromoPublish godoc
// @Summary Publica um produto em promoção
// @Description Publica um produto com promoção
// @Tags products
// @Accept json
// @Produce json
// @Param payload body service.PublishPayload true "Dados do produto em promoção"
// @Success 200 {object} PublishResponse
// @Failure 400 {object} map[string]string "JSON inválido"
// @Router /products/promo-pub [post]
func (h *ProductHandlers) PromoPublish(c *gin.Context) {
	var payload service.PublishPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}

	postID, err := h.ps.Publish(payload)
	if err != nil {
		badRequest(c, err)
		return
	}

	c.JSON(http.StatusOK, PublishResponse{PostID: postID})
}

// FollowedLastTwoWeeks godoc
// @Summary Lista produtos dos usuários seguidos
// @Description Retorna produtos publicados nos últimos 14 dias pelos usuários seguidos
// @Tags products
// @Produce json
// @Param userId path int true "ID do usuário"
// @Param order query string false "Ordenação" Enums(date_asc,date_desc)
// @Success 200 {object} FollowedPostsResponse
// @Failure 400 {object} map[string]string "Parâmetro inválido"
// @Router /products/followed/{userId}/list [get]
func (h *ProductHandlers) FollowedLastTwoWeeks(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parâmetro inválido: userId"})
		return
	}

	order := c.Query("order")
	posts, err2 := h.ps.FollowedLastTwoWeeks(userID, order)
	if err2 != nil {
		badRequest(c, err2)
		return
	}

	c.JSON(http.StatusOK, FollowedPostsResponse{
		UserID: userID,
		Posts:  posts,
	})
}

// PromoCount godoc
// @Summary Conta produtos em promoção
// @Description Retorna a quantidade de produtos em promoção de um usuário
// @Tags products
// @Produce json
// @Param user_id query int true "ID do usuário"
// @Success 200 {object} PromoCountResponse
// @Failure 400 {object} map[string]string "Parâmetro inválido"
// @Router /products/promo-pub/count [get]
func (h *ProductHandlers) PromoCount(c *gin.Context) {
	userID, err := strconv.Atoi(c.Query("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parâmetro inválido: user_id"})
		return
	}

	u, count, err2 := h.ps.PromoCount(userID)
	if err2 != nil {
		badRequest(c, err2)
		return
	}

	c.JSON(http.StatusOK, PromoCountResponse{
		UserID:             u.ID,
		UserName:           u.Name,
		PromoProductsCount: count,
	})
}

// PromoList godoc
// @Summary Lista produtos em promoção
// @Description Retorna a lista de produtos em promoção de um usuário
// @Tags products
// @Produce json
// @Param user_id query int true "ID do usuário"
// @Success 200 {object} PromoListResponse
// @Failure 400 {object} map[string]string "Parâmetro inválido"
// @Router /products/promo-pub/list [get]
func (h *ProductHandlers) PromoList(c *gin.Context) {
	userID, err := strconv.Atoi(c.Query("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parâmetro inválido: user_id"})
		return
	}
	page, limit, ok := parsePageLimit(c, 20, 100)
	if !ok {
		return
	}

	u, posts, err2 := h.ps.PromoList(userID)
	if err2 != nil {
		badRequest(c, err2)
		return
	}

	postsPage, meta := paginateSlice(posts, page, limit)
	c.JSON(http.StatusOK, gin.H{
		"user_id":   u.ID,
		"user_name": u.Name,
		"posts":     postsPage,
		"meta":      meta,
	})
}

// UploadProductImage godoc
// @Summary Upload de imagem do produto
// @Description Faz upload (multipart) e devolve a URL para usar em product.image_url
// @Tags products
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "Arquivo da imagem"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /products/me/image [post]
func (h *ProductHandlers) UploadProductImage(c *gin.Context) {
	// precisa estar autenticado
	uidAny, ok := c.Get("auth_user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token ausente"})
		return
	}
	uid := uidAny.(int)

	file, err := c.FormFile("image")
	// valida tamanho (max 4MB)
	if file.Size > 4<<20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "imagem muito grande (max 4MB)"})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "arquivo 'image' ausente"})
		return
	}

	uploadDir := "uploads/products"
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
	// nome unico: <userId>-<timestamp>.<ext>
	name := strconv.Itoa(uid) + "-" + strconv.FormatInt(time.Now().UnixNano(), 10) + ext
	localPath := filepath.Join(uploadDir, name)

	if err := c.SaveUploadedFile(file, localPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao salvar"})
		return
	}

	url := "/static/products/" + name
	c.JSON(http.StatusOK, gin.H{"image_url": url})
}

// DeleteMyPost godoc
// @Summary Apaga uma publicacao do usuario logado
// @Description Apaga uma publicacao (post) se ela pertencer ao usuario autenticado
// @Tags products
// @Produce json
// @Param postId path int true "ID da publicacao"
// @Success 204
// @Failure 400 {object} map[string]string
// @Router /products/me/{postId} [delete]
func (h *ProductHandlers) DeleteMyPost(c *gin.Context) {
	uidAny, ok := c.Get("auth_user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token ausente"})
		return
	}
	uid := uidAny.(int)

	postID, err := strconv.Atoi(c.Param("postId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parâmetro inválido: postId"})
		return
	}

	if err := h.ps.DeleteMyPost(uid, postID); err != nil {
		badRequest(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
