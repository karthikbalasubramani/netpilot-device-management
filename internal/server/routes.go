package server

const apiV1basepath = "/api/v1/"

// Register all the routes
func (server *Server) registerRoutes() {
	server.registerHealthRoutes()

	apiV1 := server.router.Group(apiV1basepath)

	server.registerV1Routes(apiV1)
}
