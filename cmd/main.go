package main

import (
	"fmt"
	"os"

	"github.com/hspedro/mcp-server-time/internal/app"
)

var (
	// Version is set by build flags
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	// Create and initialize the application
	application, err := app.New(Version, BuildTime)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize application: %v\n", err)
		os.Exit(1)
	}
	defer application.Close()

	// Run the application
	if err := application.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Application error: %v\n", err)
		os.Exit(1)
	}
}