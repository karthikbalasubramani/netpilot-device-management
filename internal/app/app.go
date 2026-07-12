package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
		return fmt.Errorf("failed to connect MongoDB: %w", err)
	}

	// Register MongoDB disconnect logic to run before application shutdown.
	defer func() {
		if err := database.Disconnect(mongoDB); err != nil {
			fmt.Println("Failed to disconnect MongoDB:", err)
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
		if err := httpServer.StartHTTPServer(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrorChan <- err
		}
	}()

	// Listen for operating system shutdown signals.
	// os.Interrupt handles Ctrl+C.
	// syscall.SIGTERM handles Docker/Kubernetes termination.
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interruptChan)

	// Keep the application running until either the server fails or a shutdown signal is received.
	select {
	case err := <-serverErrorChan:
		return fmt.Errorf("Failed to start HTTP server: %w", err)

	case <-interruptChan:
		fmt.Println("Shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := httpServer.ShutdownHTTPServer(shutdownCtx); err != nil {
			return fmt.Errorf("Failed to shutdown HTTP server gracefully: %w", err)
		}

		fmt.Println("HTTP server shutdown successfully")
		return nil
	}
}
