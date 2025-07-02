# CodeGraphGen - Knowledge Graph Generator

A powerful CLI tool for analyzing codebases and generating knowledge graphs. CodeGraphGen extracts entities and relationships from source code and stores them in various databases for analysis and visualization.

## Features

- **Multi-language Code Analysis**: Supports TypeScript, JavaScript, Python, Java, Go, Rust, C/C++, C#, PHP, Ruby, and more
- **Modular Analyzer Architecture**: Language-specific analyzers with unified interface
- **Multiple Database Backends**:
  - In-memory database for fast analysis
  - Memgraph integration for persistent storage and advanced queries
- **CLI Interface**: Easy-to-use command-line interface built with Cobra
- **Text Processing**: Extracts entities and relationships from documentation and text files
- **Statistics and Reporting**: Comprehensive analysis of code complexity and dependencies

## Supported Languages & Analysis Features

### Go Analysis

- **Packages**: Package declarations and imports
- **Structs**: Field analysis with types and tags
- **Functions**: Parameter and return type analysis
- **Methods**: Receiver type detection
- **Interfaces**: Method signature extraction
- **Types**: Type aliases and definitions
- **Constants**: Constant blocks and individual constants
- **Variables**: Global and local variable declarations

### TypeScript/JavaScript Analysis

- **Classes**: Constructor, methods, and properties
- **Functions**: Arrow functions and regular functions
- **Interfaces**: Type definitions and inheritance
- **Types**: Type aliases and union types
- **Imports/Exports**: Module dependency tracking
- **Async/Await**: Asynchronous code pattern detection

### Python Analysis

- **Classes**: Methods, properties, and inheritance
- **Functions**: Parameter and return annotations
- **Imports**: Module and package dependencies
- **Decorators**: Function and class decorators

### Java Analysis

- **Classes**: Fields, methods, and inheritance
- **Interfaces**: Method signatures
- **Packages**: Import statements and dependencies
- **Annotations**: Method and class annotations

### JSON Analysis

- **Structure**: Object hierarchy and data types
- **Schemas**: Configuration file analysis

### Example Go Analysis Output

```go
// Sample Go code
package main

import "fmt"

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

func (u *User) GetName() string {
    return u.Name
}

func main() {
    user := &User{ID: 1, Name: "John"}
    fmt.Println(user.GetName())
}
```

**Extracted Entities:**

- Package: `main`
- Import: `fmt`
- Struct: `User` with fields `ID`, `Name`
- Method: `GetName` with receiver `*User`
- Function: `main`

**Extracted Relationships:**

- `main` IMPORTS `fmt`
- `User` CONTAINS `ID`, `Name`
- `GetName` BELONGS_TO `User`
- `main` CALLS `GetName`

## Installation

### Prerequisites

- Go 1.24 or later
- (Optional) Memgraph database for persistent storage

### Build from Source

```bash
git clone <repository-url>
cd codegraphgen
go mod tidy
go build -o codegraphgen
```

### Using Go Install

```bash
go install github.com/your-org/codegraphgen@latest
```

## Usage

## Usage

CodeGraphGen provides several commands for different types of analysis:

### Available Commands

```bash
# Show help and available commands
codegraphgen --help

# Analyze a codebase directory
codegraphgen codebase [directory]

# Analyze text directly
codegraphgen text [text-to-analyze]

# Analyze a single file
codegraphgen file [file-path]

# Display knowledge graph statistics
codegraphgen stats
```

### Global Flags

- `--memgraph`: Use Memgraph database instead of in-memory storage
- `--verbose`, `-v`: Enable verbose output

### Analyze Codebase

Analyze an entire codebase directory and extract code entities and relationships:

```bash
# Analyze current directory with in-memory database
codegraphgen codebase .

# Analyze a specific project directory
codegraphgen codebase /path/to/your/project

# Analyze and store in Memgraph database
codegraphgen codebase ./my-project --memgraph

# Verbose analysis with detailed output
codegraphgen codebase . --verbose
```

### Analyze Text

Extract entities and relationships from text:

```bash
# Analyze a simple text string
codegraphgen text "This is some text to analyze"

# Analyze code snippet
codegraphgen text "function calculateSum(a, b) { return a + b; }"

# Use with Memgraph storage
codegraphgen text "Your text here" --memgraph
```

### Analyze Single File

Analyze a specific file (code or text):

```bash
# Analyze a code file
codegraphgen file ./src/main.go

# Analyze a documentation file
codegraphgen file ./README.md

# Analyze with Memgraph storage
codegraphgen file ./code.js --memgraph
```

### View Statistics

Display statistics about the knowledge graph:

```bash
# Show statistics from in-memory database
codegraphgen stats

# Show statistics from Memgraph database
codegraphgen stats --memgraph
```

### Example Output

When analyzing a codebase, you'll see output like this:

```
🔍 Analyzing codebase at: ./my-project
🧠 Using in-memory database
✅ Found 45 entities and 67 relationships

📈 Knowledge Graph Results:
Entities: 45
Relationships: 67

📊 Entity Types:
  PACKAGE: 1
  IMPORT: 8
  STRUCT: 6
  FUNCTION: 12
  METHOD: 8
  INTERFACE: 3
  TYPE: 4
  CONSTANT: 3

🔗 Relationship Types:
  IMPORTS: 8
  CONTAINS: 18
  BELONGS_TO: 15
  CALLS: 12
  DEFINES: 14

🎯 Sample Entities:
  main (PACKAGE) - map[path:.]
  fmt (IMPORT) - map[path:fmt]
  User (STRUCT) - map[fields:[ID Name] sourceFile:main.go]
  GetName (METHOD) - map[receiver:*User returnType:string sourceFile:main.go]
  main (FUNCTION) - map[parameters:[] sourceFile:main.go]

🔗 Sample Relationships:
  main -> fmt (IMPORTS)
  User -> ID (CONTAINS)
  User -> Name (CONTAINS)
  GetName -> User (BELONGS_TO)
  main -> GetName (CALLS)
```

## Architecture

### Project Structure

```
codegraphgen/
├── cmd/                    # Cobra CLI commands
│   ├── root.go            # Root command and global flags
│   ├── codebase.go        # Codebase analysis command
│   ├── text.go            # Text analysis command
│   ├── file.go            # File analysis command
│   ├── stats.go           # Statistics command
│   └── utils.go           # Shared utilities
├── internal/
│   └── core/              # Core analysis logic
│       ├── analyzer.go    # Analyzer registry
│       ├── code_processor.go        # Code analysis orchestration
│       ├── text_processor.go       # Text processing
│       ├── knowledge_graph_generator.go  # Main generator
│       ├── analyzers/     # Language-specific analyzers
│       │   ├── analyzer.go      # Analyzer interface
│       │   ├── golang.go        # Go language analyzer
│       │   ├── typescript.go    # TypeScript/JavaScript analyzer
│       │   ├── python.go        # Python analyzer
│       │   ├── java.go          # Java analyzer
│       │   ├── json.go          # JSON analyzer
│       │   └── generic.go       # Generic/fallback analyzer
│       └── graph/         # Graph types and utilities
│           └── types.go   # Entity and relationship definitions
├── db/                    # Database implementations
│   ├── inmemory.go       # In-memory database
│   └── memgraph.go       # Memgraph database connector
└── main.go               # Application entry point
```

### Core Components

1. **CLI Commands (`cmd/`)**: Cobra-based command-line interface
2. **Code Processor**: Orchestrates file analysis and language detection
3. **Analyzer Registry**: Manages language-specific analyzers
4. **Language Analyzers**: Extract entities and relationships for specific languages
5. **Database Abstraction**: Supports multiple storage backends
6. **Knowledge Graph Generator**: Coordinates the entire analysis pipeline

### Go-Specific Analysis

The Go code processor uses regular expressions and text parsing to extract:

- Package declarations: `package main`
- Import statements: `import "fmt"` or `import (...)`
- Struct definitions with field analysis
- Function signatures with parameter extraction
- Method definitions with receiver type detection
- Interface method signatures
- Type definitions and aliases
- Constant and variable declarations

### Entity Types

- `PACKAGE`: Go package
- `IMPORT`: Import statement
- `STRUCT`: Struct definition
- `FUNCTION`: Standalone function
- `METHOD`: Method with receiver
- `INTERFACE`: Interface definition
- `TYPE`: Type definition or alias
- `CONSTANT`: Constant declaration
- `FIELD`: Struct field
- `PARAMETER`: Function parameter

### Relationship Types

- `IMPORTS`: Package imports another package
- `CONTAINS`: Struct contains field, interface contains method
- `BELONGS_TO`: Method belongs to struct/type
- `DEFINES`: Package defines type/function
- `CALLS`: Function calls another function
- `IMPLEMENTS`: Type implements interface
- `USES`: General usage relationship

## Configuration

### Supported File Extensions

- **Go**: `.go`
- **TypeScript/JavaScript**: `.ts`, `.tsx`, `.js`, `.jsx`
- **Python**: `.py`
- **Java**: `.java`
- **C/C++**: `.c`, `.cpp`, `.h`, `.hpp`
- **C#**: `.cs`
- **Rust**: `.rs`
- **Ruby**: `.rb`
- **PHP**: `.php`
- **Configuration**: `.json`, `.yaml`, `.yml`, `.xml`
- **Documentation**: `.md`, `.txt`
- **Database**: `.sql`

### Directory Exclusions

The processor automatically skips common directories:

- `node_modules`
- `.git`
- `vendor`
- `target`
- `build`
- `dist`
- `.vscode`
- `.idea`

## Advanced Features

### Cypher-like Queries

The in-memory database supports basic Cypher-like queries:

```go
// Find all functions
results, err := generator.QueryKnowledgeGraph("MATCH (n:FUNCTION) RETURN n", nil)

// Find entity connections
results, err := generator.GetEntityConnections("entity-id")

// Get graph statistics
stats, err := generator.GetGraphStatistics()
```

### Codebase Metrics

The system provides comprehensive metrics:

- Total entities and relationships
- Entity type distribution
- Relationship type distribution
- File type analysis
- Code complexity indicators

## Development

### Adding New Language Support

To add support for a new programming language:

1. Add the file extension to `supportedExtensions`
2. Add language mapping in `languageMap`
3. Implement language-specific analysis method
4. Add entity extraction patterns
5. Define language-specific relationship types

### Extending Go Analysis

The Go analysis can be extended by:

- Adding more regex patterns for complex constructs
- Implementing method call analysis
- Adding cross-file reference tracking
- Supporting Go modules and dependency analysis

## Dependencies

- `github.com/google/uuid`: For generating unique entity IDs

## License

This project is part of the AI Knowledge Graph Generator suite.
