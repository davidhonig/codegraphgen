package graph

import (
	"codegraphgen/db"
	"fmt"
	"strings"
	"time"
)

// Re-export types from db package for convenience
type Properties = db.Properties
type EntityType = db.EntityType
type RelationshipType = db.RelationshipType
type Entity = db.Entity
type Relationship = db.Relationship

// Entity type constants
const (
	// Code-specific entities
	EntityTypeClass         EntityType = "CLASS"
	EntityTypeFunction      EntityType = "FUNCTION"
	EntityTypeMethod        EntityType = "METHOD"
	EntityTypeVariable      EntityType = "VARIABLE"
	EntityTypeInterface     EntityType = "INTERFACE"
	EntityTypeType          EntityType = "TYPE"
	EntityTypeModule        EntityType = "MODULE"
	EntityTypePackage       EntityType = "PACKAGE"
	EntityTypeFile          EntityType = "FILE"
	EntityTypeDirectory     EntityType = "DIRECTORY"
	EntityTypeNamespace     EntityType = "NAMESPACE"
	EntityTypeEnum          EntityType = "ENUM"
	EntityTypeConstant      EntityType = "CONSTANT"
	EntityTypeProperty      EntityType = "PROPERTY"
	EntityTypeParameter     EntityType = "PARAMETER"
	EntityTypeImport        EntityType = "IMPORT"
	EntityTypeExport        EntityType = "EXPORT"
	EntityTypeAnnotation    EntityType = "ANNOTATION"
	EntityTypeComment       EntityType = "COMMENT"
	EntityTypeTest          EntityType = "TEST"
	EntityTypeDependency    EntityType = "DEPENDENCY"
	EntityTypeAPIEndpoint   EntityType = "API_ENDPOINT"
	EntityTypeDatabaseTable EntityType = "DATABASE_TABLE"
	EntityTypeConfiguration EntityType = "CONFIGURATION"
)

// Relationship type constants
const (
	// Code-specific relationships
	RelationshipTypeInheritsFrom RelationshipType = "INHERITS_FROM"
	RelationshipTypeImplements   RelationshipType = "IMPLEMENTS"
	RelationshipTypeExtends      RelationshipType = "EXTENDS"
	RelationshipTypeCalls        RelationshipType = "CALLS"
	RelationshipTypeUses         RelationshipType = "USES"
	RelationshipTypeImports      RelationshipType = "IMPORTS"
	RelationshipTypeExports      RelationshipType = "EXPORTS"
	RelationshipTypeDependsOn    RelationshipType = "DEPENDS_ON"
	RelationshipTypeContains     RelationshipType = "CONTAINS"
	RelationshipTypeBelongsTo    RelationshipType = "BELONGS_TO"
	RelationshipTypeDefines      RelationshipType = "DEFINES"
	RelationshipTypeReferences   RelationshipType = "REFERENCES"
	RelationshipTypeOverrides    RelationshipType = "OVERRIDES"
	RelationshipTypeInstantiates RelationshipType = "INSTANTIATES"
	RelationshipTypeThrows       RelationshipType = "THROWS"
	RelationshipTypeCatches      RelationshipType = "CATCHES"
	RelationshipTypeReturns      RelationshipType = "RETURNS"
	RelationshipTypeAccepts      RelationshipType = "ACCEPTS"
	RelationshipTypeConfigures   RelationshipType = "CONFIGURES"
	RelationshipTypeTests        RelationshipType = "TESTS"
	RelationshipTypeDocuments    RelationshipType = "DOCUMENTS"
	RelationshipTypeAnnotates    RelationshipType = "ANNOTATES"
	RelationshipTypeModifies     RelationshipType = "MODIFIES"
	RelationshipTypeAccesses     RelationshipType = "ACCESSES"
	RelationshipTypeInvokes      RelationshipType = "INVOKES"
	RelationshipTypeSubscribesTo RelationshipType = "SUBSCRIBES_TO"
	RelationshipTypePublishes    RelationshipType = "PUBLISHES"
)

// KnowledgeGraph represents a complete knowledge graph
type KnowledgeGraph struct {
	Entities      []Entity       `json:"entities"`
	Relationships []Relationship `json:"relationships"`
}

// CodeFile represents a source code file
type CodeFile struct {
	Path         string    `json:"path"`
	Name         string    `json:"name"`
	Extension    string    `json:"extension"`
	Content      string    `json:"content"`
	Language     string    `json:"language"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"lastModified"`
}

// CodeEntity extends Entity with code-specific properties
type CodeEntity struct {
	Entity
	SourceFile   string   `json:"sourceFile,omitempty"`
	LineNumber   int      `json:"lineNumber,omitempty"`
	ColumnNumber int      `json:"columnNumber,omitempty"`
	SourceCode   string   `json:"sourceCode,omitempty"`
	Language     string   `json:"language,omitempty"`
	Visibility   string   `json:"visibility,omitempty"`
	IsStatic     bool     `json:"isStatic,omitempty"`
	IsAbstract   bool     `json:"isAbstract,omitempty"`
	IsAsync      bool     `json:"isAsync,omitempty"`
	Parameters   []string `json:"parameters,omitempty"`
	ReturnType   string   `json:"returnType,omitempty"`
	Complexity   int      `json:"complexity,omitempty"`
}

// CodebaseAnalysis represents analysis results of a codebase
type CodebaseAnalysis struct {
	TotalFiles        int               `json:"totalFiles"`
	TotalLines        int               `json:"totalLines"`
	Languages         map[string]int    `json:"languages"`
	FileTypes         map[string]int    `json:"fileTypes"`
	ComplexityMetrics ComplexityMetrics `json:"complexityMetrics"`
	DependencyGraph   DependencyGraph   `json:"dependencyGraph"`
}

// ComplexityMetrics represents code complexity metrics
type ComplexityMetrics struct {
	AverageComplexity float64 `json:"averageComplexity"`
	MaxComplexity     int     `json:"maxComplexity"`
	TotalComplexity   int     `json:"totalComplexity"`
}

// DependencyGraph represents dependency graph metrics
type DependencyGraph struct {
	Nodes  int `json:"nodes"`
	Edges  int `json:"edges"`
	Cycles int `json:"cycles"`
}

// generateDeterministicID generates a stable ID based on entity characteristics
func generateDeterministicID(entityType EntityType, label string, properties Properties) string {
	// Create a consistent string representation for the ID
	var keyParts []string
	keyParts = append(keyParts, strings.ToLower(string(entityType)))
	keyParts = append(keyParts, strings.ToLower(label))

	// Add path-based properties for file system entities
	if fullPath, ok := properties["fullPath"]; ok {
		keyParts = append(keyParts, strings.ToLower(fmt.Sprintf("%v", fullPath)))
	} else if path, ok := properties["path"]; ok {
		keyParts = append(keyParts, strings.ToLower(fmt.Sprintf("%v", path)))
	} else if relativePath, ok := properties["relativePath"]; ok {
		keyParts = append(keyParts, strings.ToLower(fmt.Sprintf("%v", relativePath)))
	}

	// Add source file for code entities (functions, classes, etc.)
	if sourceFile, ok := properties["sourceFile"]; ok {
		keyParts = append(keyParts, strings.ToLower(fmt.Sprintf("%v", sourceFile)))

		// Add line number for precise location
		if lineNumber, ok := properties["lineNumber"]; ok {
			keyParts = append(keyParts, fmt.Sprintf("line:%v", lineNumber))
		}
	}

	// Add namespace/package for better uniqueness
	if namespace, ok := properties["namespace"]; ok {
		keyParts = append(keyParts, strings.ToLower(fmt.Sprintf("ns:%v", namespace)))
	}
	if pkg, ok := properties["package"]; ok {
		keyParts = append(keyParts, strings.ToLower(fmt.Sprintf("pkg:%v", pkg)))
	}

	// Create hash of the combined key
	key := strings.Join(keyParts, "|")
	// hash := sha256.Sum256([]byte(key))

	// Return first 16 bytes as hex string (32 characters)
	return fmt.Sprintf("%x", key)
}

// generateDeterministicRelationshipID generates a stable ID for relationships
func generateDeterministicRelationshipID(sourceID, targetID string, relType RelationshipType) string {
	key := fmt.Sprintf("%s|%s|%s", sourceID, string(relType), targetID)
	// hash := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%x", key)
}

// graph.CreateEntity creates a new entity with a deterministic ID
func CreateEntity(label string, entityType EntityType, properties Properties) Entity {
	if properties == nil {
		properties = make(Properties)
	}

	return Entity{
		ID:         generateDeterministicID(entityType, label, properties),
		Label:      label,
		Type:       entityType,
		Properties: properties,
		Confidence: 1.0,
	}
}

// CreateRelationship creates a new relationship with a deterministic ID
func CreateRelationship(source, target string, relType RelationshipType, properties Properties) Relationship {
	if properties == nil {
		properties = make(Properties)
	}

	return Relationship{
		ID:         generateDeterministicRelationshipID(source, target, relType),
		Source:     source,
		Target:     target,
		Type:       relType,
		Properties: properties,
		Confidence: 1.0,
	}
}

// GraphStatistics represents statistics about the knowledge graph
type GraphStatistics struct {
	TotalEntities       int            `json:"totalEntities"`
	TotalRelationships  int            `json:"totalRelationships"`
	EntitiesByType      map[string]int `json:"entitiesByType"`
	RelationshipsByType map[string]int `json:"relationshipsByType"`
}
