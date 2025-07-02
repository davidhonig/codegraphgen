package db

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// MemgraphDatabase implements DatabaseConnection for Memgraph using the Neo4j driver
type MemgraphDatabase struct {
	driver   neo4j.DriverWithContext
	uri      string
	username string
	password string
}

// NewMemgraphDatabase creates a new Memgraph database connection
func NewMemgraphDatabase(uri, username, password string) *MemgraphDatabase {
	if uri == "" {
		uri = "bolt://localhost:7687" // Default Memgraph port
	}
	// Memgraph often runs without authentication in development
	if username == "" && password == "" {
		username = ""
		password = ""
	}

	return &MemgraphDatabase{
		uri:      uri,
		username: username,
		password: password,
	}
}

// Connect establishes a connection to Memgraph
func (db *MemgraphDatabase) Connect() error {
	ctx := context.Background()

	// Configure authentication
	var auth neo4j.AuthToken
	if db.username != "" || db.password != "" {
		auth = neo4j.BasicAuth(db.username, db.password, "")
	} else {
		auth = neo4j.NoAuth()
	}

	// Create driver with Memgraph-optimized configuration
	driver, err := neo4j.NewDriverWithContext(db.uri, auth, func(c *neo4j.Config) {
		c.MaxConnectionLifetime = 30 * time.Minute
		c.MaxConnectionPoolSize = 50
		c.ConnectionAcquisitionTimeout = 2 * time.Minute
		c.SocketConnectTimeout = 15 * time.Second
		c.SocketKeepalive = true
		// Note: Encryption settings may vary by Neo4j driver version
	})

	if err != nil {
		return fmt.Errorf("failed to create Memgraph driver: %w", err)
	}

	// Verify connectivity
	err = driver.VerifyConnectivity(ctx)
	if err != nil {
		driver.Close(ctx)
		return fmt.Errorf("failed to verify Memgraph connectivity: %w", err)
	}

	db.driver = driver
	log.Println("🔗 Connected to Memgraph database")

	// Optional: Check Memgraph capabilities
	if err := db.checkMemgraphCapabilities(ctx); err != nil {
		log.Printf("ℹ️ Could not check Memgraph capabilities: %v", err)
	}

	return nil
}

// Disconnect closes the connection to Memgraph
func (db *MemgraphDatabase) Disconnect() error {
	if db.driver != nil {
		ctx := context.Background()
		err := db.driver.Close(ctx)
		if err != nil {
			return fmt.Errorf("failed to close Memgraph driver: %w", err)
		}
		db.driver = nil
		log.Println("🔌 Disconnected from Memgraph database")
	}
	return nil
}

// Query executes a Cypher query against Memgraph
func (db *MemgraphDatabase) Query(cypher string, parameters Properties) ([]QueryResult, error) {
	if db.driver == nil {
		return nil, fmt.Errorf("database not connected. Call Connect() first")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Convert Properties to map[string]any for Neo4j driver
	params := make(map[string]any)
	for k, v := range parameters {
		params[k] = v
	}

	// Execute query in a read session
	session := db.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeWrite, // Memgraph supports read/write in same session
		DatabaseName: "memgraph",            // Default database name
	})
	defer session.Close(ctx)

	result, err := session.Run(ctx, cypher, params)
	if err != nil {
		log.Printf("❌ Memgraph query execution failed: %v", err)
		log.Printf("📝 Query: %s", cypher)
		log.Printf("📝 Parameters: %v", parameters)
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	// Process results
	var results []QueryResult
	for result.Next(ctx) {
		record := result.Record()
		queryResult := make(QueryResult)

		for _, key := range record.Keys {
			value, found := record.Get(key)
			if found {
				queryResult[key] = db.convertMemgraphValue(value)
			}
		}
		results = append(results, queryResult)
	}

	// Check for any errors during result processing
	if err = result.Err(); err != nil {
		return nil, fmt.Errorf("error processing query results: %w", err)
	}

	return results, nil
}

// CreateEntity creates a new entity or updates an existing one in Memgraph
func (db *MemgraphDatabase) CreateEntity(entity Entity) error {
	// Escape the entity type to handle reserved keywords
	escapedType := db.escapeLabel(string(entity.Type))
	// Escape the entity label as well for use as a node label
	escapedLabel := db.escapeLabel(entity.Label)

	// Enhanced Cypher query for entity creation/update
	// Use both the entity label and type as node labels for better organization
	cypher := fmt.Sprintf(`
		MERGE (n:%s:%s {id: $id})
		ON CREATE SET n.label = $label,
			n.confidence = $confidence,
			n.created_at = timestamp(),
			n.updated_at = timestamp()
		ON MATCH SET n.label = $label,
			n.confidence = CASE
				WHEN $confidence > n.confidence THEN $confidence
				ELSE n.confidence
			END,
			n.updated_at = timestamp()
		SET n += $properties
		RETURN n
	`, escapedType, escapedLabel)

	// Prepare properties
	params := Properties{
		"id":         entity.ID,
		"label":      entity.Label,
		"confidence": entity.Confidence,
		"properties": db.flattenProperties(entity.Properties),
	}

	_, err := db.Query(cypher, params)
	if err != nil {
		return fmt.Errorf("failed to create entity %s: %w", entity.ID, err)
	}

	return nil
}

// CreateRelationship creates a new relationship or updates an existing one in Memgraph
func (db *MemgraphDatabase) CreateRelationship(relationship Relationship) error {
	// Escape the relationship type to handle reserved keywords
	escapedType := db.escapeLabel(string(relationship.Type))

	// Enhanced Cypher query for relationship creation/update
	// Find entities by their IDs, then merge the relationship
	cypher := fmt.Sprintf(`
		MATCH (source {id: $sourceId})
		MATCH (target {id: $targetId})
		MERGE (source)-[r:%s]->(target)
		ON CREATE SET r.id = $id,
			r.confidence = $confidence,
			r.created_at = timestamp(),
			r.updated_at = timestamp()
		ON MATCH SET r.id = $id,
			r.confidence = CASE
				WHEN $confidence > r.confidence THEN $confidence
				ELSE r.confidence
			END,
			r.updated_at = timestamp()
		SET r += $properties
		RETURN r
	`, escapedType)

	// Prepare properties
	params := Properties{
		"sourceId":   relationship.Source,
		"targetId":   relationship.Target,
		"id":         relationship.ID,
		"confidence": relationship.Confidence,
		"properties": db.flattenProperties(relationship.Properties),
	}

	_, err := db.Query(cypher, params)
	if err != nil {
		return fmt.Errorf("failed to create relationship %s: %w", relationship.ID, err)
	}

	return nil
}

// CreateEntities creates multiple entities in a batch for better performance
func (db *MemgraphDatabase) CreateEntities(entities []Entity) error {
	if len(entities) == 0 {
		return nil
	}

	// For now, use individual creation as it's more reliable
	// In production, you might want to implement proper batch processing
	for _, entity := range entities {
		if err := db.CreateEntity(entity); err != nil {
			return fmt.Errorf("failed to create entity %s: %w", entity.ID, err)
		}
	}

	log.Printf("✅ Created %d entities in Memgraph", len(entities))
	return nil
}

// CreateRelationships creates multiple relationships in a batch
func (db *MemgraphDatabase) CreateRelationships(relationships []Relationship) error {
	if len(relationships) == 0 {
		return nil
	}

	// Use individual creation for relationships as UNWIND can be complex with dynamic relationship types
	for _, rel := range relationships {
		if err := db.CreateRelationship(rel); err != nil {
			return fmt.Errorf("failed to create relationship %s: %w", rel.ID, err)
		}
	}

	return nil
}

// checkMemgraphCapabilities checks available Memgraph procedures and capabilities
func (db *MemgraphDatabase) checkMemgraphCapabilities(ctx context.Context) error {
	session := db.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	// Try to get Memgraph procedures
	result, err := session.Run(ctx, "CALL mg.procedures() YIELD name RETURN count(name) as procedure_count", nil)
	if err != nil {
		return err // Not critical, just informational
	}

	if result.Next(ctx) {
		record := result.Record()
		if count, found := record.Get("procedure_count"); found {
			log.Printf("📊 Memgraph procedures available: %v", count)
		}
	}

	return nil
}

// convertMemgraphValue converts Memgraph/Neo4j driver values to Go types
func (db *MemgraphDatabase) convertMemgraphValue(value interface{}) interface{} {
	switch v := value.(type) {
	case neo4j.Node:
		return map[string]interface{}{
			"id":         v.GetId(),
			"labels":     v.Labels,
			"properties": v.Props,
		}
	case neo4j.Relationship:
		return map[string]interface{}{
			"id":         v.GetId(),
			"type":       v.Type,
			"start":      v.StartElementId,
			"end":        v.EndElementId,
			"properties": v.Props,
		}
	case neo4j.Path:
		return map[string]interface{}{
			"length":        len(v.Nodes),
			"nodes":         v.Nodes,
			"relationships": v.Relationships,
		}
	default:
		return value
	}
}

// flattenProperties flattens nested properties for Memgraph storage
func (db *MemgraphDatabase) flattenProperties(props Properties) map[string]interface{} {
	flattened := make(map[string]interface{})
	for key, value := range props {
		// Prefix properties to avoid conflicts with reserved names
		flattened[fmt.Sprintf("prop_%s", key)] = value
	}
	return flattened
}

// GetEntityByID retrieves an entity by its ID
func (db *MemgraphDatabase) GetEntityByID(id string) (*Entity, error) {
	cypher := "MATCH (n {id: $id}) RETURN n"
	params := Properties{"id": id}

	results, err := db.Query(cypher, params)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("entity not found: %s", id)
	}

	// Convert result to Entity (simplified)
	result := results[0]
	if nodeData, ok := result["n"].(map[string]interface{}); ok {
		entity := &Entity{
			ID:         id,
			Properties: make(Properties),
		}

		if label, ok := nodeData["label"].(string); ok {
			entity.Label = label
		}
		if confidence, ok := nodeData["confidence"].(float64); ok {
			entity.Confidence = confidence
		}

		return entity, nil
	}

	return nil, fmt.Errorf("invalid entity format")
}

// GetAllEntities retrieves all entities from the database
func (db *MemgraphDatabase) GetAllEntities() ([]Entity, error) {
	cypher := "MATCH (n) RETURN n LIMIT 1000" // Limit for safety
	results, err := db.Query(cypher, nil)
	if err != nil {
		return nil, err
	}

	var entities []Entity
	for _, result := range results {
		if nodeData, ok := result["n"].(map[string]interface{}); ok {
			entity := Entity{
				Properties: make(Properties),
			}

			if id, ok := nodeData["id"].(string); ok {
				entity.ID = id
			}
			if label, ok := nodeData["label"].(string); ok {
				entity.Label = label
			}
			if confidence, ok := nodeData["confidence"].(float64); ok {
				entity.Confidence = confidence
			}

			entities = append(entities, entity)
		}
	}

	return entities, nil
}

// ClearDatabase removes all nodes and relationships (useful for testing)
func (db *MemgraphDatabase) ClearDatabase() error {
	cypher := "MATCH (n) DETACH DELETE n"
	_, err := db.Query(cypher, nil)
	if err != nil {
		return fmt.Errorf("failed to clear database: %w", err)
	}
	log.Println("🗑️ Cleared Memgraph database")
	return nil
}

// escapeLabel escapes labels for Cypher queries to handle reserved keywords
func (db *MemgraphDatabase) escapeLabel(label string) string {
	// List of Memgraph/Cypher reserved keywords that need escaping
	reservedKeywords := map[string]bool{
		"DIRECTORY": true,
		"FILE":      true,
		"DATA":      true,
		"TYPE":      true,
		"INDEX":     true,
		"KEY":       true,
		"NODE":      true,
		"EDGE":      true,
		"GRAPH":     true,
		"DATABASE":  true,
		"USER":      true,
		"ROLE":      true,
		"CONFIG":    true,
		"SETTING":   true,
		"STATUS":    true,
		"VERSION":   true,
		"SESSION":   true,
		"QUERY":     true,
		"INFO":      true,
		"STATS":     true,
		"MODE":      true,
		"TIMEOUT":   true,
		"STREAM":    true,
		"TRIGGER":   true,
		"FUNCTION":  true,
		"MODULE":    true,
		"CLASS":     true,
		"METHOD":    true,
		"VARIABLE":  true,
		"CONSTANT":  true,
		"PROPERTY":  true,
		"PARAMETER": true,
		"IMPORT":    true,
		"EXPORT":    true,
		"PACKAGE":   true,
		"NAMESPACE": true,
		"INTERFACE": true,
		"ENUM":      true,
		"COMMENT":   true,
		"TEST":      true,
	}

	// Check if the label is a reserved keyword
	if reservedKeywords[strings.ToUpper(label)] {
		// Escape with backticks
		return "`" + label + "`"
	}

	// Also escape if it contains spaces or special characters
	if strings.ContainsAny(label, " -+*/()[]{}.,;:!?@#$%^&=<>|\\\"'") {
		return "`" + label + "`"
	}

	return label
}
