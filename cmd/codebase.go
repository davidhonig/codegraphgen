package cmd

import (
	"fmt"
	"log"

	"codegraphgen/db"
	"codegraphgen/internal/core"

	"github.com/spf13/cobra"
)

// codebaseCmd represents the codebase command
var codebaseCmd = &cobra.Command{
	Use:   "codebase [directory]",
	Short: "Analyze a codebase directory",
	Long: `Analyze a codebase directory and generate a knowledge graph from the source code.
This command will scan the specified directory for source code files, extract entities
(classes, functions, interfaces, etc.) and relationships (imports, inheritance, etc.),
and optionally store them in a database.

Examples:
  codegraphgen codebase .
  codegraphgen codebase ./my-project --memgraph
  codegraphgen codebase /path/to/code --memgraph`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dirPath := args[0]

		if verbose {
			fmt.Printf("🔍 Analyzing codebase at: %s\n", dirPath)
			if useMemgraph {
				fmt.Println("🔗 Using Memgraph database")
			} else {
				fmt.Println("🧠 Using in-memory database")
			}
		}

		// Initialize components
		textProcessor := core.NewTextProcessor()

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

		codeProcessor := core.NewCodeProcessor()
		generator := core.NewKnowledgeGraphGenerator(textProcessor, database)

		// Analyze the codebase
		kg, err := analyzeCodebase(codeProcessor, dirPath)
		if err != nil {
			log.Fatalf("Failed to analyze codebase: %v", err)
		}

		// Store in database
		err = generator.StoreKnowledgeGraph(kg.Entities, kg.Relationships)
		if err != nil {
			log.Fatalf("Failed to store knowledge graph: %v", err)
		}

		printKnowledgeGraph(kg)
	},
}

func init() {
	rootCmd.AddCommand(codebaseCmd)
}
