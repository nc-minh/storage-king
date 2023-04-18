package api

import (
	"github.com/gin-gonic/gin"
	"github.com/nc-minh/storage-king/utils"
)

type Server struct {
	config utils.Config
	router *gin.Engine
}

func NewServer(config utils.Config) (*Server, error) {

	server := &Server{config: config}

	server.setupRouter()

	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello, World!",
		})
	})

	// Setup routes for v1
	v1 := router.Group("/v1")
	v1.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	v1.POST("/pong", server.createDrive)

	server.router = router
}

// Start runs the HTTP server at the given address
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}
