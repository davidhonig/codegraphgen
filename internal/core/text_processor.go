package core

import (
	"codegraphgen/internal/core/graph"
	"regexp"
	"strings"
)

// TextProcessor handles text analysis for code comments and documentation
type TextProcessor struct {
	entityPatterns       map[graph.EntityType][]*regexp.Regexp
	relationshipPatterns map[graph.RelationshipType]*regexp.Regexp
}

// NewTextProcessor creates a new TextProcessor instance
func NewTextProcessor() *TextProcessor {
	tp := &TextProcessor{
		entityPatterns:       make(map[graph.EntityType][]*regexp.Regexp),
		relationshipPatterns: make(map[graph.RelationshipType]*regexp.Regexp),
	}
	tp.initializeEntityPatterns()
	tp.initializeRelationshipPatterns()
	return tp
}

// initializeEntityPatterns sets up regex patterns for entity extraction
func (tp *TextProcessor) initializeEntityPatterns() {
	// Basic code documentation patterns
	tp.entityPatterns[graph.EntityTypeComment] = []*regexp.Regexp{
		regexp.MustCompile(`/\*\*(.*?)\*/`), // JSDoc comments
		regexp.MustCompile(`/\*(.*?)\*/`),   // Block comments
		regexp.MustCompile(`//(.+)$`),       // Line comments
	}

	tp.entityPatterns[graph.EntityTypeConfiguration] = []*regexp.Regexp{
		regexp.MustCompile(`\b(config|configuration|settings|options)\b`),
		regexp.MustCompile(`\b(env|environment|ENV_\w+)\b`),
	}
}

// initializeRelationshipPatterns sets up regex patterns for relationship extraction
func (tp *TextProcessor) initializeRelationshipPatterns() {
	// Code-specific relationship patterns
	tp.relationshipPatterns[graph.RelationshipTypeImports] = regexp.MustCompile(`import\s+.*\s+from\s+['"](.+?)['"]`)
	tp.relationshipPatterns[graph.RelationshipTypeExports] = regexp.MustCompile(`export\s+(?:default\s+)?(?:class|function|const|let|var|interface|type)\s+(\w+)`)
	tp.relationshipPatterns[graph.RelationshipTypeExtends] = regexp.MustCompile(`class\s+(\w+)\s+extends\s+(\w+)`)
	tp.relationshipPatterns[graph.RelationshipTypeImplements] = regexp.MustCompile(`class\s+(\w+)\s+implements\s+(\w+)`)
	tp.relationshipPatterns[graph.RelationshipTypeUses] = regexp.MustCompile(`(\w+)\.(\w+)\(`) // Method calls
}

// ExtractEntities extracts entities from text
func (tp *TextProcessor) ExtractEntities(text string) ([]graph.Entity, error) {
	var entities []graph.Entity
	extractedTexts := make(map[string]bool)

	for entityType, patterns := range tp.entityPatterns {
		for _, pattern := range patterns {
			matches := pattern.FindAllStringSubmatch(text, -1)
			for _, match := range matches {
				if len(match) > 1 {
					entityText := strings.TrimSpace(match[1])
					if entityText != "" && !extractedTexts[entityText] && len(entityText) > 1 {
						extractedTexts[entityText] = true

						properties := graph.Properties{
							"extractedFrom": text[max(0, pattern.FindStringIndex(text)[0]-20):min(len(text), pattern.FindStringIndex(text)[1]+20)],
							"position":      pattern.FindStringIndex(text)[0],
						}

						entities = append(entities, graph.CreateEntity(entityText, entityType, properties))
					}
				}
			}
		}
	}

	return entities, nil
}

// ExtractRelationships extracts relationships from text based on entities
func (tp *TextProcessor) ExtractRelationships(text string, entities []graph.Entity) ([]graph.Relationship, error) {
	var relationships []graph.Relationship
	entityMap := make(map[string]graph.Entity)

	// Create a map for quick entity lookup
	for _, entity := range entities {
		entityMap[strings.ToLower(entity.Label)] = entity
	}

	for relationshipType, pattern := range tp.relationshipPatterns {
		matches := pattern.FindAllStringSubmatch(text, -1)

		for _, match := range matches {
			if len(match) >= 3 {
				sourceText := strings.TrimSpace(match[1])
				targetText := strings.TrimSpace(match[2])

				if sourceText == "" || targetText == "" {
					continue
				}

				sourceEntity := tp.findEntityByText(sourceText, entityMap)
				targetEntity := tp.findEntityByText(targetText, entityMap)

				if sourceEntity != nil && targetEntity != nil && sourceEntity.ID != targetEntity.ID {
					properties := graph.Properties{
						"extractedFrom": match[0],
						"confidence":    0.8,
					}

					relationships = append(relationships, graph.CreateRelationship(
						sourceEntity.ID,
						targetEntity.ID,
						relationshipType,
						properties,
					))
				}
			}
		}
	}

	return relationships, nil
}

// findEntityByText finds an entity by text with fuzzy matching
func (tp *TextProcessor) findEntityByText(text string, entityMap map[string]graph.Entity) *graph.Entity {
	lowerText := strings.ToLower(text)

	// Direct match
	if entity, exists := entityMap[lowerText]; exists {
		return &entity
	}

	// Partial match
	for key, entity := range entityMap {
		if strings.Contains(key, lowerText) || strings.Contains(lowerText, key) {
			return &entity
		}
	}

	return nil
}

// ProcessText processes text and extracts entities and relationships
func (tp *TextProcessor) ProcessText(text string) ([]graph.Entity, []graph.Relationship, error) {
	entities, err := tp.ExtractEntities(text)
	if err != nil {
		return nil, nil, err
	}

	relationships, err := tp.ExtractRelationships(text, entities)
	if err != nil {
		return entities, nil, err
	}

	return entities, relationships, nil
}

// ProcessCodeText processes code text with source file context
func (tp *TextProcessor) ProcessCodeText(text string, sourceFile string) ([]graph.Entity, []graph.Relationship, error) {
	entities, relationships, err := tp.ProcessText(text)
	if err != nil {
		return entities, relationships, err
	}

	// Add source file context if provided
	if sourceFile != "" {
		for i := range entities {
			entities[i].Properties["sourceFile"] = sourceFile
		}
		for i := range relationships {
			relationships[i].Properties["sourceFile"] = sourceFile
		}
	}

	return entities, relationships, nil
}

// CleanText normalizes text for processing
func (tp *TextProcessor) CleanText(text string) string {
	// Normalize whitespace
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	// Remove special chars except basic punctuation
	text = regexp.MustCompile(`[^\w\s.,!?;:()\-]`).ReplaceAllString(text, "")
	return strings.TrimSpace(text)
}

// SplitIntoSentences splits text into sentences for better processing
func (tp *TextProcessor) SplitIntoSentences(text string) []string {
	sentences := regexp.MustCompile(`[.!?]+`).Split(text, -1)
	var result []string

	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if len(sentence) > 0 {
			result = append(result, sentence)
		}
	}

	return result
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
