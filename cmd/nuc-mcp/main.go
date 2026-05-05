package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/marstid/nuc/internal/mcp"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	transport := flag.String("transport", "stdio", "Transport mode: stdio or http")
	addr := flag.String("addr", "localhost:8080", "Listen address for HTTP transport")
	apiKey := flag.String("api-key", "", "Nucleus API key")
	baseURL := flag.String("base-url", "", "Nucleus API base URL")
	flag.Parse()

	cfg, err := mcp.Resolve(*apiKey, *baseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	srv, err := mcp.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	switch *transport {
	case "stdio":
		if err := srv.Run(ctx, &mcpsdk.StdioTransport{}); err != nil {
			log.Printf("Server failed: %v", err)
		}
	case "http":
		handler := mcpsdk.NewStreamableHTTPHandler(func(*http.Request) *mcpsdk.Server {
			return srv
		}, nil)
		log.Printf("MCP server listening on %s", *addr)
		//nolint:gosec // G114: MCP server; timeouts handled by MCP SDK transport
		if err := http.ListenAndServe(*addr, handler); err != nil {
			log.Fatal(err)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown transport: %s (use stdio or http)\n", *transport)
		os.Exit(1)
	}
}
