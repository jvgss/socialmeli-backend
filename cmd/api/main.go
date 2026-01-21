package main

import (
	"os"
	"time"

	_ "socialmeli/docs"
	"socialmeli/internal/http"
	"socialmeli/internal/service"
	"socialmeli/internal/store"

	"github.com/gin-contrib/cors"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           SocialMeli API
// @version         1.0
// @description     API REST para gerenciamento de rede social de vendas (SocialMeli)
// @host            localhost:8080
// @BasePath        /
// @schemes         http
func main() {
	// Se tiver DATABASE_URL, usa Postgres (SQLStore). Senão, usa MemoryStore.
	dsn := os.Getenv("DATABASE_URL")

	if dsn != "" {
		println("✅ USANDO SQLStore")
	} else {
		println("⚠️ USANDO MemoryStore")
	}

	var st store.Store
	if dsn != "" {
		sqlSt, err := store.NewSQLStore(dsn)
		if err != nil {
			panic(err)
		}
		st = sqlSt
	} else {
		mem := store.NewMemoryStore()
		store.SeedDefault(mem)
		st = mem
	}

	us := service.NewUserService(st)
	ps := service.NewProductService(st)

	as := service.NewAuthService(st)

	r := http.NewRouter(us, ps, as)

	r.SetTrustedProxies(nil)

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5173",
			"http://127.0.0.1:5173",
		},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders: []string{
			"Content-Length",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	_ = r.Run(":" + port)
}
