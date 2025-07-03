package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"codegraphgen/pkg/rest"

	"github.com/spf13/cobra"
)

var (
	port int
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the CodeGraphGen web server",
	Long: `Start a web server that provides REST API endpoints for analyzing code and managing knowledge graphs.

The server provides the following endpoints:
  POST /api/analyze/text     - Analyze text content
  POST /api/analyze/file     - Analyze a file
  POST /api/analyze/codebase - Analyze a codebase directory
  GET  /api/stats            - Get knowledge graph statistics
  GET  /api/entities         - Get all entities
  GET  /api/relationships    - Get all relationships
  GET  /api/query            - Execute a query against the graph
  GET  /health               - Health check endpoint
  GET  /                     - API documentation

Examples:
  codegraphgen server
  codegraphgen server --port 8080 --memgraph
  codegraphgen server --verbose --port 3000`,
	Run: func(cmd *cobra.Command, args []string) {
		if verbose {
			fmt.Printf("🚀 Starting CodeGraphGen server on port %d\n", port)
			if useMemgraph {
				fmt.Println("🔗 Using Memgraph database")
			} else {
				fmt.Println("🧠 Using in-memory database")
			}
		}

		// Create server configuration
		config := rest.Config{
			Port:        port,
			Verbose:     verbose,
			UseMemgraph: useMemgraph,
		}

		// Create and start server
		srv, err := rest.NewServer(config)
		if err != nil {
			log.Fatalf("Failed to create server: %v", err)
		}

		// Setup graceful shutdown
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		go func() {
			<-c
			fmt.Println("\n🔄 Shutting down server...")
			if err := srv.Shutdown(); err != nil {
				log.Printf("Error during shutdown: %v", err)
			}
			os.Exit(0)
		}()

		// Start server
		if verbose {
			fmt.Printf("📡 Server listening on http://localhost:%d\n", port)
			fmt.Printf("📖 API documentation available at http://localhost:%d/\n", port)
			fmt.Printf("❤️  Health check at http://localhost:%d/health\n", port)
		}

		if err := srv.Start(); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to run the server on")
}
