package analyzers

import (
	"codegraphgen/internal/core/graph"
	"regexp"
	"strings"
)

// PythonAnalyzer implements the LanguageAnalyzer interface for Python
type PythonAnalyzer struct{}

func (pa *PythonAnalyzer) Name() string                 { return "Python Analyzer" }
func (pa *PythonAnalyzer) SupportedLanguages() []string { return []string{"python"} }
func (pa *PythonAnalyzer) Analyze(file graph.CodeFile, fileEntity graph.Entity) ([]graph.Entity, []graph.Relationship, error) {
	return analyzePythonFile(file, fileEntity)
}

// analyzePythonFile analyzes a Python source file
func analyzePythonFile(file graph.CodeFile, fileEntity graph.Entity) ([]graph.Entity, []graph.Relationship, error) {
	entities := []graph.Entity{fileEntity}
	var relationships []graph.Relationship

	content := file.Content

	// Extract Python classes
	classRegex := regexp.MustCompile(`^class\s+(\w+)(?:\(([^)]*)\))?:`)
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)

		if match := classRegex.FindStringSubmatch(line); len(match) > 1 {
			className := match[1]
			inheritance := match[2]

			var extends []string
			if inheritance != "" {
				extends = strings.Split(inheritance, ",")
				for j, ext := range extends {
					extends[j] = strings.TrimSpace(ext)
				}
			}

			classEntity := graph.CreateEntity(className, graph.EntityTypeClass, graph.Properties{
				"sourceFile": file.Path,
				"lineNumber": i + 1,
				"language":   "python",
				"extends":    extends,
			})
			entities = append(entities, classEntity)
			relationships = append(relationships, graph.CreateRelationship(
				fileEntity.ID, classEntity.ID, graph.RelationshipTypeDefines, nil))
		}
	}

	// Extract Python functions
	funcRegex := regexp.MustCompile(`^def\s+(\w+)\s*\(`)
	methodRegex := regexp.MustCompile(`^\s+def\s+(\w+)\s*\(`)

	for i, line := range lines {
		// Top-level functions
		if match := funcRegex.FindStringSubmatch(line); len(match) > 1 {
			funcName := match[1]
			funcEntity := graph.CreateEntity(funcName, graph.EntityTypeFunction, graph.Properties{
				"sourceFile": file.Path,
				"lineNumber": i + 1,
				"language":   "python",
			})
			entities = append(entities, funcEntity)
			relationships = append(relationships, graph.CreateRelationship(
				fileEntity.ID, funcEntity.ID, graph.RelationshipTypeDefines, nil))
		}

		// Methods (indented functions)
		if match := methodRegex.FindStringSubmatch(line); len(match) > 1 {
			methodName := match[1]
			methodEntity := graph.CreateEntity(methodName, graph.EntityTypeMethod, graph.Properties{
				"sourceFile": file.Path,
				"lineNumber": i + 1,
				"language":   "python",
			})
			entities = append(entities, methodEntity)
			// Note: In a full implementation, you'd associate methods with their classes
		}
	}

	// Extract imports
	importRegex := regexp.MustCompile(`^(?:from\s+(\S+)\s+)?import\s+(.+)`)
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if match := importRegex.FindStringSubmatch(line); len(match) > 2 {
			source := match[1]
			imports := match[2]

			if source == "" {
				source = imports
			}

			importEntity := graph.CreateEntity(imports, graph.EntityTypeImport, graph.Properties{
				"source":     source,
				"lineNumber": i + 1,
				"language":   "python",
			})
			entities = append(entities, importEntity)
			relationships = append(relationships, graph.CreateRelationship(
				fileEntity.ID, importEntity.ID, graph.RelationshipTypeImports, nil))
		}
	}

	return entities, relationships, nil
}
