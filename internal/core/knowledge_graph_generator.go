package core

import (
	"codegraphgen/db"
	"codegraphgen/internal/core/graph"
	"fmt"
	"log"
	"os"
	"strings"
)

// KnowledgeGraphGenerator handles knowledge graph generation and management
type KnowledgeGraphGenerator struct {
	textProcessor *TextProcessor
	database      db.DatabaseConnection
}

// NewKnowledgeGraphGenerator creates a new KnowledgeGraphGenerator instance
func NewKnowledgeGraphGenerator(textProcessor *TextProcessor, database db.DatabaseConnection) *KnowledgeGraphGenerator {
	return &KnowledgeGraphGenerator{
		textProcessor: textProcessor,
		database:      database,
	}
}

// ExtractEntitiesFromText extracts entities from text
func (kg *KnowledgeGraphGenerator) ExtractEntitiesFromText(text string) ([]graph.Entity, error) {
	cleanText := kg.textProcessor.CleanText(text)
	return kg.textProcessor.ExtractEntities(cleanText)
}

// ExtractRelationshipsFromText extracts relationships from text
func (kg *KnowledgeGraphGenerator) ExtractRelationshipsFromText(text string, entities []graph.Entity) ([]graph.Relationship, error) {
	cleanText := kg.textProcessor.CleanText(text)
	return kg.textProcessor.ExtractRelationships(cleanText, entities)
}

// GenerateKnowledgeGraph generates a knowledge graph from text
func (kg *KnowledgeGraphGenerator) GenerateKnowledgeGraph(text string) (*graph.KnowledgeGraph, error) {
	fmt.Println("🔍 Extracting entities and relationships...")

	entities, relationships, err := kg.textProcessor.ProcessCodeText(text, "")
	if err != nil {
		return nil, fmt.Errorf("failed to process text: %w", err)
	}

	fmt.Printf("✅ Extracted %d entities and %d relationships\n", len(entities), len(relationships))

	return &graph.KnowledgeGraph{
		Entities:      entities,
		Relationships: relationships,
	}, nil
}

// StoreKnowledgeGraph stores entities and relationships in the database
// Entities are updated if they already exist, relationships are merged
func (kg *KnowledgeGraphGenerator) StoreKnowledgeGraph(entities []graph.Entity, relationships []graph.Relationship) error {
	fmt.Println("💾 Storing knowledge graph in database...")

	// Store/update entities first
	for i, entity := range entities {
		if err := kg.database.CreateEntity(entity); err != nil {
			return fmt.Errorf("failed to create/update entity %s: %w", entity.Label, err)
		}
		if (i+1)%10 == 0 {
			fmt.Printf("📊 Processed %d/%d entities\n", i+1, len(entities))
		}
	}

	fmt.Printf("✅ Stored/updated %d entities\n", len(entities))

	// Then store/merge relationships
	successfulRelationships := 0
	for i, relationship := range relationships {
		if err := kg.database.CreateRelationship(relationship); err != nil {
			log.Printf("⚠️ Failed to create relationship %s->%s (%s): %v",
				relationship.Source, relationship.Target, relationship.Type, err)
		} else {
			successfulRelationships++
		}
		if (i+1)%10 == 0 {
			fmt.Printf("📊 Processed %d/%d relationships\n", i+1, len(relationships))
		}
	}

	fmt.Printf("✅ Successfully stored %d/%d relationships\n", successfulRelationships, len(relationships))
	fmt.Println("✅ Knowledge graph stored successfully")

	// Debug: Check if functions have relationships
	if err := kg.debugFunctionRelationships(); err != nil {
		log.Printf("⚠️ Debug check failed: %v", err)
	}

	return nil
}

// debugFunctionRelationships checks if function entities have relationships (for debugging)
func (kg *KnowledgeGraphGenerator) debugFunctionRelationships() error {
	// Find all function entities
	functions, err := kg.QueryKnowledgeGraph(`
		MATCH (f:FUNCTION)
		RETURN f.id as id, f.label as label
		LIMIT 5
	`, nil)
	if err != nil {
		return fmt.Errorf("failed to query functions: %w", err)
	}

	fmt.Printf("🔍 Found %d function entities for debugging\n", len(functions))

	for _, fn := range functions {
		if id, ok := fn["id"].(string); ok {
			if label, ok := fn["label"].(string); ok {
				// Check relationships for this function
				rels, err := kg.QueryKnowledgeGraph(`
					MATCH (f {id: $id})-[r]-(other)
					RETURN type(r) as relType, labels(other) as otherLabels, other.label as otherLabel
				`, graph.Properties{"id": id})
				if err != nil {
					log.Printf("⚠️ Failed to query relationships for function %s: %v", label, err)
					continue
				}

				fmt.Printf("🔗 Function '%s' has %d relationships:\n", label, len(rels))
				for _, rel := range rels {
					if relType, ok := rel["relType"].(string); ok {
						if otherLabel, ok := rel["otherLabel"].(string); ok {
							fmt.Printf("  - %s -> %s\n", relType, otherLabel)
						}
					}
				}
			}
		}
	}

	return nil
}

// QueryKnowledgeGraph executes a query against the knowledge graph
func (kg *KnowledgeGraphGenerator) QueryKnowledgeGraph(cypher string, parameters graph.Properties) ([]db.QueryResult, error) {
	return kg.database.Query(cypher, parameters)
}

// ProcessTextFile processes a text file and generates a knowledge graph
func (kg *KnowledgeGraphGenerator) ProcessTextFile(filePath string) (*graph.KnowledgeGraph, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	return kg.GenerateKnowledgeGraph(string(content))
}

// ProcessMultipleTexts processes multiple texts and combines them into one knowledge graph
func (kg *KnowledgeGraphGenerator) ProcessMultipleTexts(texts []string) (*graph.KnowledgeGraph, error) {
	var allEntities []graph.Entity
	var allRelationships []graph.Relationship

	for _, text := range texts {
		kgraph, err := kg.GenerateKnowledgeGraph(text)
		if err != nil {
			return nil, fmt.Errorf("failed to process text: %w", err)
		}

		allEntities = append(allEntities, kgraph.Entities...)
		allRelationships = append(allRelationships, kgraph.Relationships...)
	}

	// Deduplicate entities based on label and type
	uniqueEntities := kg.deduplicateEntities(allEntities)

	return &graph.KnowledgeGraph{
		Entities:      uniqueEntities,
		Relationships: allRelationships,
	}, nil
}

// GetEntityConnections gets all connections for a specific entity
func (kg *KnowledgeGraphGenerator) GetEntityConnections(entityID string) ([]db.QueryResult, error) {
	cypher := `
		MATCH (e {id: $entityId})-[r]-(connected)
		RETURN e, r, connected
	`
	parameters := graph.Properties{"entityId": entityID}
	return kg.QueryKnowledgeGraph(cypher, parameters)
}

// FindEntitiesByType finds all entities of a specific type
func (kg *KnowledgeGraphGenerator) FindEntitiesByType(entityType string) ([]db.QueryResult, error) {
	cypher := fmt.Sprintf("MATCH (n:%s) RETURN n", entityType)
	return kg.QueryKnowledgeGraph(cypher, nil)
}

// GetGraphStatistics returns statistics about the knowledge graph
func (kg *KnowledgeGraphGenerator) GetGraphStatistics() (*graph.GraphStatistics, error) {
	entityStats, err := kg.QueryKnowledgeGraph(`
		MATCH (n)
		RETURN labels(n)[0] as type, count(*) as count
	`, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity stats: %w", err)
	}

	relationshipStats, err := kg.QueryKnowledgeGraph(`
		MATCH ()-[r]->()
		RETURN type(r) as type, count(*) as count
	`, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get relationship stats: %w", err)
	}

	entitiesByType := make(map[string]int)
	relationshipsByType := make(map[string]int)

	for _, stat := range entityStats {
		if entityType, ok := stat["type"].(string); ok {
			if count, ok := stat["count"].(int); ok {
				entitiesByType[entityType] = count
			}
		}
	}

	for _, stat := range relationshipStats {
		if relType, ok := stat["type"].(string); ok {
			if count, ok := stat["count"].(int); ok {
				relationshipsByType[relType] = count
			}
		}
	}

	totalEntities := 0
	for _, count := range entitiesByType {
		totalEntities += count
	}

	totalRelationships := 0
	for _, count := range relationshipsByType {
		totalRelationships += count
	}

	return &graph.GraphStatistics{
		TotalEntities:       totalEntities,
		TotalRelationships:  totalRelationships,
		EntitiesByType:      entitiesByType,
		RelationshipsByType: relationshipsByType,
	}, nil
}

// ExportKnowledgeGraph exports the complete knowledge graph
func (kg *KnowledgeGraphGenerator) ExportKnowledgeGraph() (*graph.KnowledgeGraph, error) {
	entitiesResult, err := kg.QueryKnowledgeGraph("MATCH (n) RETURN n", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to export entities: %w", err)
	}

	relationshipsResult, err := kg.QueryKnowledgeGraph("MATCH (a)-[r]->(b) RETURN a, r, b", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to export relationships: %w", err)
	}

	// This is a simplified export - in practice you'd want to properly reconstruct the entities and relationships
	var entities []graph.Entity
	var relationships []graph.Relationship

	for _, result := range entitiesResult {
		if entity, ok := result["n"].(graph.Entity); ok {
			entities = append(entities, entity)
		}
	}

	for _, result := range relationshipsResult {
		if relationship, ok := result["r"].(graph.Relationship); ok {
			relationships = append(relationships, relationship)
		}
	}

	return &graph.KnowledgeGraph{
		Entities:      entities,
		Relationships: relationships,
	}, nil
}

// ClearDatabase clears all data from the database
func (kg *KnowledgeGraphGenerator) ClearDatabase() error {
	_, err := kg.database.Query("MATCH (n) DETACH DELETE n", nil)
	if err != nil {
		return fmt.Errorf("failed to clear database: %w", err)
	}
	fmt.Println("🧹 Database cleared")
	return nil
}

// deduplicateEntities removes duplicate entities based on label and type
func (kg *KnowledgeGraphGenerator) deduplicateEntities(entities []graph.Entity) []graph.Entity {
	seen := make(map[string]bool)
	var unique []graph.Entity

	for _, entity := range entities {
		key := fmt.Sprintf("%s-%s", strings.ToLower(entity.Label), entity.Type)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, entity)
		}
	}

	return unique
}

// Advanced querying methods

// FindPathBetweenEntities finds paths between two entities
func (kg *KnowledgeGraphGenerator) FindPathBetweenEntities(fromLabel, toLabel string) ([]db.QueryResult, error) {
	cypher := `
		MATCH (from {label: $fromLabel}), (to {label: $toLabel})
		MATCH path = shortestPath((from)-[*]-(to))
		RETURN path, length(path) as pathLength
		ORDER BY pathLength
		LIMIT 5
	`
	parameters := graph.Properties{
		"fromLabel": fromLabel,
		"toLabel":   toLabel,
	}
	return kg.QueryKnowledgeGraph(cypher, parameters)
}

// FindInfluentialEntities finds entities with the most connections
func (kg *KnowledgeGraphGenerator) FindInfluentialEntities(limit int) ([]db.QueryResult, error) {
	cypher := `
		MATCH (n)-[r]-()
		WITH n, count(r) as connections
		RETURN n, connections
		ORDER BY connections DESC
		LIMIT $limit
	`
	parameters := graph.Properties{"limit": limit}
	return kg.QueryKnowledgeGraph(cypher, parameters)
}

// FindSimilarEntities finds entities similar to a given entity
func (kg *KnowledgeGraphGenerator) FindSimilarEntities(entityID string, limit int) ([]db.QueryResult, error) {
	cypher := `
		MATCH (target {id: $entityId})-[r1]-(common)-[r2]-(similar)
		WHERE target <> similar
		WITH similar, count(common) as commonConnections
		RETURN similar, commonConnections
		ORDER BY commonConnections DESC
		LIMIT $limit
	`
	parameters := graph.Properties{
		"entityId": entityID,
		"limit":    limit,
	}
	return kg.QueryKnowledgeGraph(cypher, parameters)
}
