package main

import (
	"SneakerFlash/internal/server"
)

func main() {
	r := server.NewHttpServer()
	r.Run()
}
