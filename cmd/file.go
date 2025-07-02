package cmd

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"codegraphgen/db"
	"codegraphgen/internal/core"
	"codegraphgen/internal/core/graph"

	"github.com/spf13/cobra"
)

// fileCmd represents the file command
var fileCmd = &cobra.Command{
	Use:   "file [file-path]",
	Short: "Analyze a text file and generate a knowledge graph",
	Long: `Analyze a text file and generate a knowledge graph from its contents.
This command reads the specified file and extracts entities and relationships.

Examples:
  codegraphgen file ./document.txt
  codegraphgen file ./code.js
  codegraphgen file ./README.md`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]

		if verbose {
			fmt.Printf("📄 Analyzing file: %s\n", filePath)
		}

		// Initialize components
		textProcessor := core.NewTextProcessor()
		codeProcessor := core.NewCodeProcessor()

		var database db.DatabaseConnection
		if useMemgraph {
			memgraphDB := db.NewMemgraphDatabase("bolt://localhost:7687", "", "")
			if err := memgraphDB.Connect(); err != nil {
				log.Fatalf("Failed to connect to Memgraph: %v", err)
			}
			database = memgraphDB
			defer memgraphDB.Disconnect()
		} else {
			database = db.NewInMemoryDatabase()
			if err := database.Connect(); err != nil {
				log.Fatalf("Failed to connect to in-memory database: %v", err)
			}
		}

		generator := core.NewKnowledgeGraphGenerator(textProcessor, database)

		// Determine if this is a code file and process accordingly
		var kg *graph.KnowledgeGraph
		var err error

		if isCodeFile(filePath) {
			// Process as a code file
			if verbose {
				fmt.Println("🔍 Extracting entities and relationships...")
			}

			entities, relationships, err := codeProcessor.ProcessSingleFile(filePath)
			if err != nil {
				log.Fatalf("Failed to process code file: %v", err)
			}

			// Store in database
			if err := generator.StoreKnowledgeGraph(entities, relationships); err != nil {
				log.Fatalf("Failed to store knowledge graph: %v", err)
			}

			kg = &graph.KnowledgeGraph{
				Entities:      entities,
				Relationships: relationships,
			}
		} else {
			// Process as a text file
			kg, err = generator.ProcessTextFile(filePath)
			if err != nil {
				log.Fatalf("Failed to process text file: %v", err)
			}
		}

		printKnowledgeGraph(kg)
	},
}

func init() {
	rootCmd.AddCommand(fileCmd)
}

// isCodeFile determines if a file is a source code file based on its extension
func isCodeFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	codeExtensions := map[string]bool{
		".go":    true,
		".js":    true,
		".ts":    true,
		".tsx":   true,
		".jsx":   true,
		".py":    true,
		".java":  true,
		".cpp":   true,
		".c":     true,
		".h":     true,
		".hpp":   true,
		".cs":    true,
		".php":   true,
		".rb":    true,
		".rs":    true,
		".swift": true,
		".kt":    true,
		".scala": true,
		".json":  true,
	}
	return codeExtensions[ext]
}
