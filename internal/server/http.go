package server

import (
	"SneakerFlash/internal/db"
	"SneakerFlash/internal/handler"
	"SneakerFlash/internal/repository"
	"SneakerFlash/internal/service"

	"github.com/gin-gonic/gin"
)

func NewHttpServer() *gin.Engine {
	// repo 层
	userRepo := repository.NewUserRepo(db.DB)
	productRepo := repository.NewProductRepo(db.DB)
	orderRepo := repository.NewOrderRepo(db.DB)

	// service 层
	userServicer := service.NewUserService(userRepo)
	productServicer := service.NewProductService(productRepo)
	orderServicer := service.NewOrderService(orderRepo)

	// handler 层
	userHandler := handler.NewUserHandler(userServicer)
	productHandler := handler.NewProductHandler(productServicer)
	orderHandler := handler.NewOrderHandler(orderServicer)

	r := gin.Default()
	r.Group("/api")
	{

	}
	return r
}
