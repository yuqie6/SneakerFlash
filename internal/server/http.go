package server

import "github.com/gin-gonic/gin"

func NewHttpServer() *gin.Engine {
	r := gin.Default()
	r.Group("/api")
	{

	}
	return r
}
