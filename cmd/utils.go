package cmd

import (
	"encoding/json"
	"fmt"
	"log"

	"codegraphgen/internal/core"
	"codegraphgen/internal/core/graph"
)

// analyzeCodebase analyzes a codebase directory and returns a knowledge graph
func analyzeCodebase(processor *core.CodeProcessor, dirPath string) (*graph.KnowledgeGraph, error) {
	fmt.Printf("🔍 Analyzing codebase at: %s\n", dirPath)

	entities, relationships, err := processor.AnalyzeCodebase(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to process directory: %w", err)
	}

	fmt.Printf("✅ Found %d entities and %d relationships\n", len(entities), len(relationships))

	return &graph.KnowledgeGraph{
		Entities:      entities,
		Relationships: relationships,
	}, nil
}

// printKnowledgeGraph prints a summary of the knowledge graph
func printKnowledgeGraph(kg *graph.KnowledgeGraph) {
	fmt.Println("\n📈 Knowledge Graph Results:")
	fmt.Printf("Entities: %d\n", len(kg.Entities))
	fmt.Printf("Relationships: %d\n", len(kg.Relationships))

	// Group entities by type
	entityCounts := make(map[graph.EntityType]int)
	for _, entity := range kg.Entities {
		entityCounts[entity.Type]++
	}

	fmt.Println("\n📊 Entity Types:")
	for entityType, count := range entityCounts {
		fmt.Printf("  %s: %d\n", entityType, count)
	}

	// Group relationships by type
	relationshipCounts := make(map[graph.RelationshipType]int)
	for _, rel := range kg.Relationships {
		relationshipCounts[rel.Type]++
	}

	fmt.Println("\n🔗 Relationship Types:")
	for relType, count := range relationshipCounts {
		fmt.Printf("  %s: %d\n", relType, count)
	}

	// Show first few entities as examples
	fmt.Println("\n🎯 Sample Entities:")
	for i, entity := range kg.Entities {
		if i >= 5 { // Show only first 5 entities
			break
		}
		fmt.Printf("  %s (%s) - %s\n", entity.Label, entity.Type, entity.Properties)
	}

	// Show first few relationships as examples
	fmt.Println("\n🔗 Sample Relationships:")
	for i, rel := range kg.Relationships {
		if i >= 5 { // Show only first 5 relationships
			break
		}

		// Find source and target entity labels
		var sourceLabel, targetLabel string
		for _, entity := range kg.Entities {
			if entity.ID == rel.Source {
				sourceLabel = entity.Label
			}
			if entity.ID == rel.Target {
				targetLabel = entity.Label
			}
		}

		fmt.Printf("  %s -> %s (%s)\n", sourceLabel, targetLabel, rel.Type)
	}
}

// printStats prints knowledge graph statistics
func printStats(stats *graph.GraphStatistics) {
	fmt.Println("\n📊 Knowledge Graph Statistics:")
	fmt.Printf("Total Entities: %d\n", stats.TotalEntities)
	fmt.Printf("Total Relationships: %d\n", stats.TotalRelationships)

	fmt.Println("\nEntities by Type:")
	for entityType, count := range stats.EntitiesByType {
		fmt.Printf("  %s: %d\n", entityType, count)
	}

	fmt.Println("\nRelationships by Type:")
	for relType, count := range stats.RelationshipsByType {
		fmt.Printf("  %s: %d\n", relType, count)
	}
}

// Helper function to pretty print JSON
func printJSON(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Printf("Error marshalling JSON: %v", err)
		return
	}
	fmt.Println(string(b))
}
