package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/server"

	mcpserver "github.com/rafaribe/portabase-mcp/internal/mcp"
	"github.com/rafaribe/portabase-mcp/internal/portabase"
)

func main() {
	baseURL := os.Getenv("PORTABASE_BASE_URL")
	token := os.Getenv("PORTABASE_API_TOKEN")
	mode := os.Getenv("PORTABASE_TRANSPORT") // "sse" or "stdio" (default)

	if baseURL == "" || token == "" {
		log.Fatal("PORTABASE_BASE_URL and PORTABASE_API_TOKEN are required")
	}

	client := portabase.NewClient(baseURL, token)
	s := mcpserver.NewServer(client)

	switch mode {
	case "sse":
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		sseServer := server.NewSSEServer(s)
		log.Printf("Starting SSE server on :%s", port)
		if err := sseServer.Start(fmt.Sprintf(":%s", port)); err != nil {
			log.Fatal(err)
		}
	default:
		if err := server.ServeStdio(s); err != nil {
			log.Fatal(err)
		}
	}
}
