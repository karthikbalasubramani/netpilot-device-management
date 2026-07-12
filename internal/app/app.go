package app

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/karthikbalasubramani/netpilot-device-management/internal/config"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/database"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/server"
)

// Run initializes application dependencies and starts the NetPilot API server.
func Run() error {
	// Load application configuration from environment variables.
	cfg := config.Load()

	fmt.Println("NetPilot API started successfully")
	fmt.Println("Application Name: ", cfg.AppName)
	fmt.Println("Environment: ", cfg.AppEnv)
	fmt.Println("Port: ", cfg.AppPort)

	// Establish MongoDB connection during application startup.
	mongoDB, err := database.ConnectMongoDB(cfg)
	if err != nil {
		return fmt.Errorf("Failed to connect MongoDB: %w", err)
	}

	// Register MongoDB disconnect logic to run before application shutdown.
	defer func() {
		if err := database.Disconnect(mongoDB); err != nil {
			fmt.Println("Failed to disconnect MongoDB: %w", err)
		} else {
			fmt.Println("MongoDB Disconnected Successfully")
		}
	}()

	fmt.Println("Mongo Database Connected Successfully")

	// Initialize HTTP server with application configuration.
	httpServer := server.NewHTTPServer(cfg)

	// Start the HTTP server in a separate goroutine so the main flow can listen for shutdown signals.
	serverErrorChan := make(chan error, 1)

	go func() {
		if err := httpServer.StartHTTPServer(); err != nil {
			serverErrorChan <- err
		}
	}()

	// Listen for operating system interrupt signal, such as Ctrl+C.
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt)

	// Keep the application running until either the server fails or a shutdown signal is received.
	select {
	case err := <-serverErrorChan:
		return fmt.Errorf("failed to start HTTP server: %w", err)

	case <-interruptChan:
		fmt.Println("Shutdown signal received")
		return nil
	}
}
