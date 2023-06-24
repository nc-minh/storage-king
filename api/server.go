package api

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/nc-minh/storage-king/db/sqlc"
	"github.com/nc-minh/storage-king/google"
	"github.com/nc-minh/storage-king/templates"
	"github.com/nc-minh/storage-king/utils"
	cors "github.com/rs/cors/wrapper/gin"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

type Server struct {
	config       utils.Config
	router       *gin.Engine
	oauth2Config *oauth2.Config
	store        db.Store
	google       google.GoogleAuthService
}

func NewServer(config utils.Config, oauth2Config *oauth2.Config, store db.Store, google google.GoogleAuthService) (*Server, error) {

	server := &Server{config: config, oauth2Config: oauth2Config, store: store, google: google}

	server.setupRouter()

	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()
	router.Use(cors.AllowAll())
	router.Static("/public", "./public")

	router.GET("/", func(c *gin.Context) {
		ts, err := template.ParseFiles(templates.VIEWS.Home)
		if err != nil {
			log.Print(err.Error())
			http.Error(c.Writer, "Internal Server Error", 500)
			return
		}

		err = ts.Execute(c.Writer, nil)
		if err != nil {
			log.Print(err.Error())
			http.Error(c.Writer, "Internal Server Error", 500)
		}
	})

	router.GET("/authenticate", func(c *gin.Context) {
		ts, err := template.ParseFiles(templates.VIEWS.Authenticate)
		if err != nil {
			log.Print(err.Error())
			http.Error(c.Writer, "Internal Server Error", 500)
			return
		}

		err = ts.Execute(c.Writer, nil)
		if err != nil {
			log.Print(err.Error())
			http.Error(c.Writer, "Internal Server Error", 500)
		}
	})

	router.GET("/dashboard", server.authMiddleware, func(c *gin.Context) {
		ts, err := template.ParseFiles(templates.VIEWS.Dashboard)
		if err != nil {
			log.Print(err.Error())
			http.Error(c.Writer, "Internal Server Error", 500)
			return
		}

		err = ts.Execute(c.Writer, nil)
		if err != nil {
			log.Print(err.Error())
			http.Error(c.Writer, "Internal Server Error", 500)
		}
	})

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

	v1.POST("/upload", server.accessTokenMiddleware, server.upload)

	v1.GET("/auth-url", server.createAuthURL)
	v1.POST("/google/refresh-token", server.refreshToken)

	router.GET("/auth/google/callback", server.createStorage)

	server.router = router
}

// Start runs the HTTP server at the given address
func (server *Server) Start(address string) error {
	log.Info().Msg("starting HTTP server")
	return server.router.Run(address)
}

func (server *Server) HttpLogger() gin.IRoutes {
	return server.router.Use(utils.HttpLogger())
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
