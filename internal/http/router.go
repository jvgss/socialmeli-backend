package http

import (
	"os"
	"time"

	"socialmeli/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter(us *service.UserService, ps *service.ProductService, as *service.AuthService) *gin.Engine {
	r := gin.Default()
	r.MaxMultipartMemory = 8 << 20 // 8MB

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5173",
			"http://127.0.0.1:5173",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// static para avatar
	_ = os.MkdirAll("uploads/avatars", 0o755)
	r.Static("/static/avatars", "uploads/avatars")

	// static para imagens de produtos
	_ = os.MkdirAll("uploads/products", 0o755)
	r.Static("/static/products", "uploads/products")

	uh := NewUserHandlers(us)
	uc := NewUsersCatalogHandlers(us)
	ph := NewProductHandlers(ps)
	ah := NewAuthHandlers(as)
	prof := NewProfileHandlers(us)

	// auth
	r.POST("/auth/register", ah.Register)
	r.POST("/auth/login", ah.Login)

	authed := r.Group("/", AuthMiddleware())
	authed.GET("/auth/me", prof.Me)
	authed.GET("/users/me/posts", prof.MyPosts)
	authed.POST("/users/me/avatar", prof.UploadAvatar)
	// upload de imagem de produto (usa Bearer token)
	authed.POST("/products/me/image", ph.UploadProductImage)
	// apagar publicacao do usuario logado
	authed.DELETE("/products/me/:postId", ph.DeleteMyPost)

	// health
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	// users catalog
	r.GET("/users", uc.List)
	r.POST("/users", uc.Create)

	// follow
	r.POST("/users/:userId/follow/:userIdToFollow", uh.Follow)
	r.POST("/users/:userId/unfollow/:userIdToUnfollow", uh.Unfollow)
	// social lists
	r.GET("/users/:userId/followers/count", uh.FollowersCount)
	r.GET("/users/:userId/followers/list", uh.FollowersList)
	r.GET("/users/:userId/followed/list", uh.FollowedList)

	// profile publico
	r.GET("/users/:userId/profile", prof.GetProfile)

	// products
	r.POST("/products/publish", ph.Publish)
	r.GET("/products/followed/:userId/list", ph.FollowedLastTwoWeeks)

	r.POST("/products/promo-pub", ph.PromoPublish)
	r.GET("/products/promo-pub/count", ph.PromoCount)
	r.GET("/products/promo-pub/list", ph.PromoList)

	return r
}
