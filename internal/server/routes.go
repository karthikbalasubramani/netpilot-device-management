package server

import "github.com/karthikbalasubramani/netpilot-device-management/internal/middleware"

const ApiV1basepath = "/api/v1/"

// Register all the routes
func (server *Server) registerRoutes() {
	server.registerHealthRoutes()

	publicV1 := server.router.Group(
		ApiV1basepath,
	)
	server.registerV1Routes(publicV1)

	protectedV1 := server.router.Group(
		ApiV1basepath,
	)
	protectedV1.Use(
		middleware.Authentication(
			server.accessTokenVerifier,
		),
	)
	server.registerProtectedV1Routes(
		protectedV1,
	)
}
