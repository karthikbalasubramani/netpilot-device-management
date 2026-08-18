package server

const ApiV1basepath = "/api/v1/"

// Register all the routes
func (server *Server) registerRoutes() {
	server.registerHealthRoutes()

	apiV1 := server.router.Group(ApiV1basepath)

	server.registerV1Routes(apiV1)
}
