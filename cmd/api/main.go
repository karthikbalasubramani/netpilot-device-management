package main

import (
	"log"

	"github.com/karthikbalasubramani/netpilot-device-management/internal/app"
)

func main() {
	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
