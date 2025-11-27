package server

import (
	"SneakerFlash/internal/config"
	"SneakerFlash/internal/db"
	"SneakerFlash/internal/handler"
	"SneakerFlash/internal/infra/redis"
	"SneakerFlash/internal/middlerware"
	"SneakerFlash/internal/repository"
	"SneakerFlash/internal/service"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewHttpServer() *gin.Engine {
	// repo 层
	userRepo := repository.NewUserRepo(db.DB)
	productRepo := repository.NewProductRepo(db.DB)

	// service 层
	userServicer := service.NewUserService(userRepo)
	productServicer := service.NewProductService(productRepo)
	seckillServicer := service.NewSeckillService(db.DB, productRepo)
	orderServicer := service.NewOrderService(db.DB, productRepo, userRepo)
	uploadServicer := service.NewUploadService(config.Conf.Server.UploadDir)
	vipServicer := service.NewVIPService(db.DB, userRepo)
	couponServicer := service.NewCouponService(db.DB)

	// handler 层
	userHandler := handler.NewUserHandler(userServicer)
	productHandler := handler.NewProductHandler(productServicer)
	seckillHandler := handler.NewSeckillHandler(seckillServicer)
	orderHandler := handler.NewOrderHandler(orderServicer)
	uploadHandler := handler.NewUploadHandler(uploadServicer)
	vipHandler := handler.NewVIPHandler(vipServicer)
	couponHandler := handler.NewCouponHandler(couponServicer, vipServicer)

	// 注册路由
	r := gin.New()
	r.Use(middlerware.SlogMiddlerware(), middlerware.MetricsMiddleware(), middlerware.SlogRecovery())
	if config.Conf.Risk.Enable {
		r.Use(middlerware.BlackListMiddleware(redis.RDB))
		r.Use(middlerware.GrayListMiddleware(redis.RDB))
	}
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
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/metrics", func(c *gin.Context) {
		middlerware.MetricsHandler(c)
	})

	// 处理预检请求
	r.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	api := r.Group("/api/v1")
	{
		api.POST("/register", userHandler.Register)
		if config.Conf.Risk.Enable {
			loginLimit := middlerware.InterfaceLimiter(redis.RDB, middlerware.BuildLimit(config.Conf.Risk.LoginRate, "rl:login", 60), "登录过于频繁，请稍后再试")
			api.POST("/login", loginLimit, userHandler.Login)
		} else {
			api.POST("/login", userHandler.Login)
		}
		api.POST("/refresh", userHandler.Refresh)

		api.GET("/products", productHandler.ListProducts)
		api.GET("/product/:id", productHandler.GetProduct)

		// 支付回调（示例）
		if config.Conf.Risk.Enable {
			payLimit := middlerware.InterfaceLimiter(redis.RDB, middlerware.BuildLimit(config.Conf.Risk.PayRate, "rl:pay", 60), "支付请求过于频繁")
			api.POST("/payment/callback", payLimit, orderHandler.PaymentCallback)
		} else {
			api.POST("/payment/callback", orderHandler.PaymentCallback)
		}
	}

	auth := api.Group("/")
	auth.Use(middlerware.JWTauth())
	{
		auth.GET("/profile", userHandler.GetProfile)
		auth.PUT("/profile", userHandler.UpdateProfile)
		auth.POST("/upload", uploadHandler.UploadImage)
		auth.GET("/vip/profile", vipHandler.GetProfile)
		auth.POST("/vip/purchase", vipHandler.Purchase)
		auth.GET("/coupons/mine", couponHandler.ListMyCoupons)
		auth.POST("/coupons/purchase", couponHandler.PurchaseCoupon)

		auth.POST("/products", productHandler.Create)
		auth.PUT("/products/:id", productHandler.UpdateProduct)
		auth.DELETE("/products/:id", productHandler.DeleteProduct)
		auth.GET("/products/mine", productHandler.ListMyProducts)

		if config.Conf.Risk.Enable {
			seckillLimit := middlerware.InterfaceLimiter(redis.RDB, middlerware.BuildLimit(config.Conf.Risk.SeckillRate, "rl:seckill", 30), "秒杀过于频繁，请稍后再试")
			paramLimit := middlerware.ParamLimiter(redis.RDB, middlerware.BuildLimit(config.Conf.Risk.ProductRate, "rl:hot:product", 30), "product_id", "该商品访问过于频繁，请稍后再试")
			auth.POST("/seckill", seckillLimit, paramLimit, seckillHandler.Seckill)
		} else {
			auth.POST("/seckill", seckillHandler.Seckill)
		}

		auth.GET("/orders", orderHandler.ListOrders)
		auth.GET("/orders/:id", orderHandler.GetOrder)
		auth.GET("/orders/poll/:order_num", orderHandler.PollOrder)
		auth.POST("/orders/:id/apply-coupon", orderHandler.ApplyCoupon)
	}
	return r
}
