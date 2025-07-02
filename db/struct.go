package db

// Properties is a map of property key-value pairs
type Properties map[string]interface{}

// QueryResult represents a query result
type QueryResult map[string]interface{}

// EntityType represents the type of an entity
type EntityType string

// RelationshipType represents the type of a relationship
type RelationshipType string

// Entity represents a knowledge graph entity
type Entity struct {
	ID         string     `json:"id"`
	Label      string     `json:"label"`
	Type       EntityType `json:"type"`
	Properties Properties `json:"properties"`
	Confidence float64    `json:"confidence,omitempty"`
}

// Relationship represents a knowledge graph relationship
type Relationship struct {
	ID         string           `json:"id"`
	Source     string           `json:"source"` // Entity ID
	Target     string           `json:"target"` // Entity ID
	Type       RelationshipType `json:"type"`
	Properties Properties       `json:"properties"`
	Confidence float64          `json:"confidence,omitempty"`
}

// DatabaseConnection interface defines database operations
type DatabaseConnection interface {
	Connect() error
	Disconnect() error
	Query(cypher string, parameters Properties) ([]QueryResult, error)
	CreateEntity(entity Entity) error
	CreateRelationship(relationship Relationship) error
}

