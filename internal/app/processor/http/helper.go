package http

import (
	"log"

	"github.com/gin-gonic/gin"
)

func regRoute(router gin.IRouter, method, path string, handler gin.HandlerFunc, name string) {
	switch method {
	case "GET":
		router.GET(path, handler)
	case "POST":
		router.POST(path, handler)
	case "PUT":
		router.PUT(path, handler)
	case "PATCH":
		router.PATCH(path, handler)
	case "DELETE":
		router.DELETE(path, handler)
	}
	log.Printf("Registered route: name=%s method=%s path=%s", name, method, path)
}

func logRoutes(router *gin.Engine) {
	routes := router.Routes()
	log.Println("All registered routes:")
	for _, route := range routes {
		log.Printf("  %s %s", route.Method, route.Path)
	}
}
