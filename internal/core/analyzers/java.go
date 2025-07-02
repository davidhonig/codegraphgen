package analyzers

import (
	"codegraphgen/internal/core/graph"
	"path/filepath"
	"regexp"
	"strings"
)

// JavaAnalyzer implements the LanguageAnalyzer interface for Java
type JavaAnalyzer struct{}

func (ja *JavaAnalyzer) Name() string                 { return "Java Analyzer" }
func (ja *JavaAnalyzer) SupportedLanguages() []string { return []string{"java"} }
func (ja *JavaAnalyzer) Analyze(file graph.CodeFile, fileEntity graph.Entity) ([]graph.Entity, []graph.Relationship, error) {
	return analyzeJavaFile(file, fileEntity)
}

// analyzeJavaFile analyzes a Java source file
func analyzeJavaFile(file graph.CodeFile, fileEntity graph.Entity) ([]graph.Entity, []graph.Relationship, error) {
	entities := []graph.Entity{fileEntity}
	var relationships []graph.Relationship

	content := file.Content
	lines := strings.Split(content, "\n")

	// Extract package declaration
	packageRegex := regexp.MustCompile(`package\s+([^;]+);`)
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if match := packageRegex.FindStringSubmatch(line); len(match) > 1 {
			packageEntity := graph.CreateEntity(match[1], graph.EntityTypePackage, graph.Properties{
				"sourceFile": file.Path,
				"lineNumber": i + 1,
				"language":   "java",
			})
			entities = append(entities, packageEntity)
			relationships = append(relationships, graph.CreateRelationship(
				fileEntity.ID, packageEntity.ID, graph.RelationshipTypeDefines, nil))
		}
	}

	// Extract imports
	importRegex := regexp.MustCompile(`import\s+(?:static\s+)?([^;]+);`)
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if match := importRegex.FindStringSubmatch(line); len(match) > 1 {
			importPath := match[1]
			importName := filepath.Base(strings.Replace(importPath, ".", "/", -1))
			importEntity := graph.CreateEntity(importName, graph.EntityTypeImport, graph.Properties{
				"source":     importPath,
				"lineNumber": i + 1,
				"language":   "java",
			})
			entities = append(entities, importEntity)
			relationships = append(relationships, graph.CreateRelationship(
				fileEntity.ID, importEntity.ID, graph.RelationshipTypeImports, nil))
		}
	}

	// Extract classes
	classRegex := regexp.MustCompile(`(?:public\s+|private\s+|protected\s+)?(?:abstract\s+)?(?:final\s+)?class\s+(\w+)(?:\s+extends\s+(\w+))?(?:\s+implements\s+(.+?))?`)
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if match := classRegex.FindStringSubmatch(line); len(match) > 1 {
			className := match[1]
			extends := match[2]
			implements := match[3]

			var extendsSlice []string
			if extends != "" {
				extendsSlice = []string{extends}
			}

			var implementsSlice []string
			if implements != "" {
				implementsSlice = strings.Split(implements, ",")
				for j, impl := range implementsSlice {
					implementsSlice[j] = strings.TrimSpace(impl)
				}
			}

			classEntity := graph.CreateEntity(className, graph.EntityTypeClass, graph.Properties{
				"sourceFile": file.Path,
				"lineNumber": i + 1,
				"language":   "java",
				"isPublic":   strings.Contains(line, "public"),
				"isAbstract": strings.Contains(line, "abstract"),
				"extends":    extendsSlice,
				"implements": implementsSlice,
			})
			entities = append(entities, classEntity)
			relationships = append(relationships, graph.CreateRelationship(
				fileEntity.ID, classEntity.ID, graph.RelationshipTypeDefines, nil))
		}
	}

	// Extract methods (simplified)
	methodRegex := regexp.MustCompile(`(?:public|private|protected)\s+(?:static\s+)?(?:final\s+)?(\w+)\s+(\w+)\s*\(`)
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if match := methodRegex.FindStringSubmatch(line); len(match) > 2 {
			returnType := match[1]
			methodName := match[2]

			methodEntity := graph.CreateEntity(methodName, graph.EntityTypeMethod, graph.Properties{
				"sourceFile": file.Path,
				"lineNumber": i + 1,
				"language":   "java",
				"returnType": returnType,
				"isPublic":   strings.Contains(line, "public"),
				"isStatic":   strings.Contains(line, "static"),
			})
			entities = append(entities, methodEntity)
			// Note: In a full implementation, you'd associate methods with their classes
		}
	}

	return entities, relationships, nil
}
