package cmd

import (
	"fmt"
	"log"

	"codegraphgen/db"
	"codegraphgen/internal/core"

	"github.com/spf13/cobra"
)

// statsCmd represents the stats command
var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Display knowledge graph statistics",
	Long: `Display statistics about the knowledge graph stored in the database.
This shows information about entities, relationships, and their types.

Examples:
  codegraphgen stats
  codegraphgen stats --memgraph`,
	Run: func(cmd *cobra.Command, args []string) {
		if verbose {
			fmt.Println("📊 Getting knowledge graph statistics")
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

		stats, err := generator.GetGraphStatistics()
		if err != nil {
			log.Fatalf("Failed to get statistics: %v", err)
		}

		printStats(stats)
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
