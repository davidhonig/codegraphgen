package db

import (
	"fmt"
	"log"
	"sync"
)

// InMemoryDatabase is a simple in-memory implementation of DatabaseConnection
type InMemoryDatabase struct {
	entities      map[string]Entity
	relationships map[string]Relationship
	mutex         sync.RWMutex
}

// NewInMemoryDatabase creates a new in-memory database
func NewInMemoryDatabase() *InMemoryDatabase {
	return &InMemoryDatabase{
		entities:      make(map[string]Entity),
		relationships: make(map[string]Relationship),
	}
}

// Connect establishes a connection (no-op for in-memory)
func (db *InMemoryDatabase) Connect() error {
	log.Println("🔗 Connected to in-memory database")
	return nil
}

// Disconnect closes the connection (no-op for in-memory)
func (db *InMemoryDatabase) Disconnect() error {
	log.Println("🔌 Disconnected from in-memory database")
	return nil
}

// Query executes a query against the in-memory database
func (db *InMemoryDatabase) Query(cypher string, parameters Properties) ([]QueryResult, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	// This is a simplified query implementation
	// In a real implementation, you'd need a proper Cypher parser

	if cypher == "MATCH (n) RETURN n" {
		results := make([]QueryResult, 0, len(db.entities))
		for _, entity := range db.entities {
			results = append(results, QueryResult{"n": entity})
		}
		return results, nil
	}

	if cypher == "MATCH (a)-[r]->(b) RETURN a, r, b" {
		results := make([]QueryResult, 0, len(db.relationships))
		for _, rel := range db.relationships {
			sourceEntity, sourceExists := db.entities[rel.Source]
			targetEntity, targetExists := db.entities[rel.Target]

			if sourceExists && targetExists {
				result := QueryResult{
					"a": sourceEntity,
					"r": rel,
					"b": targetEntity,
				}
				results = append(results, result)
			}
		}
		return results, nil
	}

	// Handle basic entity type queries
	if len(cypher) > 12 && cypher[:12] == "MATCH (n:" {
		// Extract entity type from query like "MATCH (n:CLASS) RETURN n"
		endIdx := -1
		for i := 12; i < len(cypher); i++ {
			if cypher[i] == ')' {
				endIdx = i
				break
			}
		}

		if endIdx != -1 {
			entityType := cypher[12:endIdx]
			results := make([]QueryResult, 0)

			for _, entity := range db.entities {
				if string(entity.Type) == entityType {
					results = append(results, QueryResult{"n": entity})
				}
			}
			return results, nil
		}
	}

	// Handle statistics queries
	if cypher == `
		MATCH (n)
		RETURN labels(n)[0] as type, count(*) as count
	` {
		typeCounts := make(map[string]int)
		for _, entity := range db.entities {
			typeCounts[string(entity.Type)]++
		}

		results := make([]QueryResult, 0, len(typeCounts))
		for entityType, count := range typeCounts {
			results = append(results, QueryResult{
				"type":  entityType,
				"count": count,
			})
		}
		return results, nil
	}

	if cypher == `
		MATCH ()-[r]->()
		RETURN type(r) as type, count(*) as count
	` {
		typeCounts := make(map[string]int)
		for _, relationship := range db.relationships {
			typeCounts[string(relationship.Type)]++
		}

		results := make([]QueryResult, 0, len(typeCounts))
		for relType, count := range typeCounts {
			results = append(results, QueryResult{
				"type":  relType,
				"count": count,
			})
		}
		return results, nil
	}

	log.Printf("⚠️ Unsupported query: %s", cypher)
	return []QueryResult{}, nil
}

// CreateEntity creates a new entity in the database
// CreateEntity creates a new entity or updates an existing one in the database
func (db *InMemoryDatabase) CreateEntity(entity Entity) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if existingEntity, exists := db.entities[entity.ID]; exists {
		// Update existing entity - keep higher confidence and merge properties
		updatedEntity := existingEntity
		updatedEntity.Label = entity.Label
		if entity.Confidence > existingEntity.Confidence {
			updatedEntity.Confidence = entity.Confidence
		}

		// Merge properties
		if updatedEntity.Properties == nil {
			updatedEntity.Properties = make(Properties)
		}
		for k, v := range entity.Properties {
			updatedEntity.Properties[k] = v
		}

		db.entities[entity.ID] = updatedEntity
		log.Printf("🔄 Updated entity: %s (%s)", updatedEntity.Label, updatedEntity.Type)
	} else {
		// Create new entity
		db.entities[entity.ID] = entity
		log.Printf("✅ Created entity: %s (%s)", entity.Label, entity.Type)
	}
	return nil
}

// CreateRelationship creates a new relationship or updates an existing one in the database
func (db *InMemoryDatabase) CreateRelationship(relationship Relationship) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	// Check if source and target entities exist
	if _, sourceExists := db.entities[relationship.Source]; !sourceExists {
		return fmt.Errorf("source entity %s not found", relationship.Source)
	}

	if _, targetExists := db.entities[relationship.Target]; !targetExists {
		return fmt.Errorf("target entity %s not found", relationship.Target)
	}

	// Check for existing relationship between same entities with same type
	var existingID string
	for id, rel := range db.relationships {
		if rel.Source == relationship.Source &&
			rel.Target == relationship.Target &&
			rel.Type == relationship.Type {
			existingID = id
			break
		}
	}

	if existingID != "" {
		// Update existing relationship
		existingRel := db.relationships[existingID]
		if relationship.Confidence > existingRel.Confidence {
			existingRel.Confidence = relationship.Confidence
		}

		// Merge properties
		if existingRel.Properties == nil {
			existingRel.Properties = make(Properties)
		}
		for k, v := range relationship.Properties {
			existingRel.Properties[k] = v
		}

		db.relationships[existingID] = existingRel
		log.Printf("🔄 Updated relationship: %s -[%s]-> %s",
			db.entities[relationship.Source].Label,
			relationship.Type,
			db.entities[relationship.Target].Label)
	} else {
		// Create new relationship
		db.relationships[relationship.ID] = relationship
		log.Printf("✅ Created relationship: %s -[%s]-> %s",
			db.entities[relationship.Source].Label,
			relationship.Type,
			db.entities[relationship.Target].Label)
	}
	return nil
}

// CreateEntities creates multiple entities in batch
func (db *InMemoryDatabase) CreateEntities(entities []Entity) error {
	for _, entity := range entities {
		if err := db.CreateEntity(entity); err != nil {
			return err
		}
	}
	return nil
}

// CreateRelationships creates multiple relationships in batch
func (db *InMemoryDatabase) CreateRelationships(relationships []Relationship) error {
	for _, relationship := range relationships {
		if err := db.CreateRelationship(relationship); err != nil {
			return err
		}
	}
	return nil
}

// GetEntityByID returns an entity by its ID
func (db *InMemoryDatabase) GetEntityByID(id string) (*Entity, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	if entity, exists := db.entities[id]; exists {
		return &entity, nil
	}
	return nil, fmt.Errorf("entity not found")
}

// GetAllEntities returns all entities
func (db *InMemoryDatabase) GetAllEntities() []Entity {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	entities := make([]Entity, 0, len(db.entities))
	for _, entity := range db.entities {
		entities = append(entities, entity)
	}
	return entities
}

// ClearDatabase removes all nodes and relationships (useful for testing)
func (db *InMemoryDatabase) ClearDatabase() error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	db.entities = make(map[string]Entity)
	db.relationships = make(map[string]Relationship)
	log.Println("🗑️ Cleared in-memory database")
	return nil
}
