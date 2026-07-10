package app

import (
	"fmt"

	"github.com/karthikbalasubramani/netpilot-device-management/internal/config"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/server"
)

func Run() error {
	cfg := config.Load()
	fmt.Println("NetPilot API started successfully")
	fmt.Println("Application Name: ", cfg.AppName)
	fmt.Println("Environment: ", cfg.AppEnv)
	fmt.Println("Port: ", cfg.AppPort)

	httpServer := server.NewHTTPServer(cfg)

	err := httpServer.StartHTTPServer()
	if err != nil {
		return fmt.Errorf("Failed to start HTTP Server: ", err)
	}
	return nil
}
