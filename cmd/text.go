package cmd

import (
	"fmt"
	"log"

	"codegraphgen/db"
	"codegraphgen/internal/core"

	"github.com/spf13/cobra"
)

// textCmd represents the text command
var textCmd = &cobra.Command{
	Use:   "text [text-to-analyze]",
	Short: "Analyze text and generate a knowledge graph",
	Long: `Analyze a piece of text and generate a knowledge graph from it.
This command extracts entities and relationships from the provided text.

Examples:
  codegraphgen text "This is some text to analyze"
  codegraphgen text "function calculateSum(a, b) { return a + b; }"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		text := args[0]

		if verbose {
			fmt.Printf("📝 Analyzing text: %s\n", text)
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

		generator := core.NewKnowledgeGraphGenerator(textProcessor, database)

		kg, err := generator.GenerateKnowledgeGraph(text)
		if err != nil {
			log.Fatalf("Failed to generate knowledge graph: %v", err)
		}

		printKnowledgeGraph(kg)
	},
}

func init() {
	rootCmd.AddCommand(textCmd)
}
