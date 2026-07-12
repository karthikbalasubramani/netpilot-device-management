package main

import (
	"log"

	"github.com/karthikbalasubramani/netpilot-device-management/internal/app"
)

// main is the application entry point.
// It starts the NetPilot API and terminates the application if startup fails.
func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
