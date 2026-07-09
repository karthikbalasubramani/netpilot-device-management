package app

import (
	"fmt"

	"github.com/karthikbalasubramani/netpilot-device-management/internal/config"
)

func Run() {
	cfg := config.Load()
	fmt.Println("NetPilot API started successfully")
	fmt.Println("Application Name: ", cfg.AppName)
	fmt.Println("Environment: ", cfg.AppEnv)
	fmt.Println("Port: ", cfg.AppPort)
}
