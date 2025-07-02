package analyzers

import (
	"codegraphgen/internal/core/graph"
	"encoding/json"
)

// JSONAnalyzer implements the LanguageAnalyzer interface for JSON
type JSONAnalyzer struct{}

func (ja *JSONAnalyzer) Name() string                 { return "JSON Analyzer" }
func (ja *JSONAnalyzer) SupportedLanguages() []string { return []string{"json"} }
func (ja *JSONAnalyzer) Analyze(file graph.CodeFile, fileEntity graph.Entity) ([]graph.Entity, []graph.Relationship, error) {
	return analyzeJSONFile(file, fileEntity)
}

// analyzeJSONFile analyzes a JSON file for dependencies
func analyzeJSONFile(file graph.CodeFile, fileEntity graph.Entity) ([]graph.Entity, []graph.Relationship, error) {
	entities := []graph.Entity{fileEntity}
	var relationships []graph.Relationship

	// For package.json, extract dependencies
	if file.Name == "package.json" {
		var packageData map[string]interface{}
		if err := json.Unmarshal([]byte(file.Content), &packageData); err == nil {
			// Extract dependencies
			if deps, ok := packageData["dependencies"].(map[string]interface{}); ok {
				for name, version := range deps {
					if versionStr, ok := version.(string); ok {
						depEntity := graph.CreateEntity(name, graph.EntityTypeDependency, graph.Properties{
							"version":    versionStr,
							"sourceFile": file.Path,
							"type":       "dependency",
						})
						entities = append(entities, depEntity)
						relationships = append(relationships, graph.CreateRelationship(
							fileEntity.ID, depEntity.ID, graph.RelationshipTypeDependsOn, nil))
					}
				}
			}

			// Extract devDependencies
			if devDeps, ok := packageData["devDependencies"].(map[string]interface{}); ok {
				for name, version := range devDeps {
					if versionStr, ok := version.(string); ok {
						depEntity := graph.CreateEntity(name, graph.EntityTypeDependency, graph.Properties{
							"version":    versionStr,
							"sourceFile": file.Path,
							"type":       "devDependency",
						})
						entities = append(entities, depEntity)
						relationships = append(relationships, graph.CreateRelationship(
							fileEntity.ID, depEntity.ID, graph.RelationshipTypeDependsOn, nil))
					}
				}
			}
		}
	}

	return entities, relationships, nil
}
