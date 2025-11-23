package server

import (
	"SneakerFlash/internal/config"
	"SneakerFlash/internal/db"
	"SneakerFlash/internal/handler"
	"SneakerFlash/internal/middlerware"
	"SneakerFlash/internal/repository"
	"SneakerFlash/internal/service"
	"net/http"

	"github.com/gin-contrib/cors"
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
	orderServicer := service.NewOrderService(db.DB)
	uploadServicer := service.NewUploadService(config.Conf.Server.UploadDir)

	// handler 层
	userHandler := handler.NewUserHandler(userServicer)
	productHandler := handler.NewProductHandler(productServicer)
	seckillHandler := handler.NewSeckillHandler(seckillServicer)
	orderHandler := handler.NewOrderHandler(orderServicer, productServicer)
	uploadHandler := handler.NewUploadHandler(uploadServicer)

	// 注册路由
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			// 预留按需放通其他前端域名
			return true
		},
		AllowWebSockets: true,
	}))
	uploadDir := config.Conf.Server.UploadDir
	if uploadDir == "" {
		uploadDir = "uploads"
	}
	r.Static("/uploads", uploadDir)

	// 处理预检请求
	r.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	api := r.Group("/api/v1")
	{
		api.POST("/register", userHandler.Register)
		api.POST("/login", userHandler.Login)
		api.POST("/refresh", userHandler.Refresh)

		api.GET("/products", productHandler.ListProducts)
		api.GET("/product/:id", productHandler.GetProduct)

		// 支付回调（示例）
		api.POST("/payment/callback", orderHandler.PaymentCallback)
	}

	auth := api.Group("/")
	auth.Use(middlerware.JWTauth())
	{
		auth.GET("/profile", userHandler.GetProfile)
		auth.PUT("/profile", userHandler.UpdateProfile)
		auth.POST("/upload", uploadHandler.UploadImage)

		auth.POST("/products", productHandler.Create)

		auth.POST("/seckill", seckillHandler.Seckill)

		auth.POST("/orders", orderHandler.CreateOrder)
		auth.GET("/orders", orderHandler.ListOrders)
		auth.GET("/orders/:id", orderHandler.GetOrder)
	}
	return r
}
