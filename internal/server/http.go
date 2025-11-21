package server

import (
	"SneakerFlash/internal/db"
	"SneakerFlash/internal/handler"
	"SneakerFlash/internal/middlerware"
	"SneakerFlash/internal/repository"
	"SneakerFlash/internal/service"

	"github.com/gin-gonic/gin"
)

func NewHttpServer() *gin.Engine {
	// repo 层
	userRepo := repository.NewUserRepo(db.DB)
	productRepo := repository.NewProductRepo(db.DB)

	// service 层
	userServicer := service.NewUserService(userRepo)
	productServicer := service.NewProductService(productRepo)
	seckillServicer := service.NewSeckillService()

	// handler 层
	userHandler := handler.NewUserHandler(userServicer)
	productHandler := handler.NewProductHandler(productServicer)
	seckillHandler := handler.NewSeckillHandler(seckillServicer)

	// 注册路由
	r := gin.Default()
	api := r.Group("/api/v1")
	{
		api.POST("/register", userHandler.Register)
		api.POST("/login", userHandler.Login)

		api.GET("/products", productHandler.ListProducts)
		api.GET("/product/:id", productHandler.GetProduct)
	}

	auth := api.Group("/")
	auth.Use(middlerware.JWTauth())
	{
		auth.GET("/profile", userHandler.GetProfile)

		auth.POST("/products", productHandler.Create)

		auth.POST("/seckill", seckillHandler.Seckill)
	}
	return r
}
