package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/config"
)

// Struct Server consists of Config and Router
type Server struct {
	config *config.Config
	router *gin.Engine
}

// NewHTTPServer initializes new HTTP Server Instance
func NewHTTPServer(cfg *config.Config) *Server {
	server := &Server{
		config: cfg,
		router: gin.Default(),
	}

	server.registerRoutes()

	return server
}

// Starts HTTP Server
func (s *Server) StartHTTPServer() error {
	address := fmt.Sprintf(":%s", s.config.AppPort)
	return s.router.Run(address)
}

// registerRoutes will register all the routes
func (s *Server) registerRoutes() {
	s.router.GET("/health", s.healthCheck)
}

// healthCheck returns the current application health status.
func (s *Server) healthCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status":        "ok",
		"service":       s.config.AppName,
		"environment":   s.config.AppEnv,
		"server_status": "Running",
	})
}
