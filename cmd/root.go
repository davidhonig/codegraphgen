package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	useMemgraph bool
	verbose     bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "codegraphgen",
	Short: "A code knowledge graph generator",
	Long: `CodeGraphGen is a CLI tool for analyzing codebases and generating knowledge graphs.
It can extract entities and relationships from source code and store them in various databases.

Examples:
  codegraphgen server
  codegraphgen codebase ./my-project
  codegraphgen codebase . --memgraph
  codegraphgen text "your text here"
  codegraphgen file ./document.txt
  codegraphgen stats`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVar(&useMemgraph, "memgraph", false, "Use Memgraph database instead of in-memory")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
}
