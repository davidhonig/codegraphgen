package rest

import (
	"fmt"
	"net/http"

	"codegraphgen/db"
	"codegraphgen/internal/core"
	"codegraphgen/internal/core/graph"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Server represents the REST API server
type Server struct {
	generator     *core.KnowledgeGraphGenerator
	codeProcessor *core.CodeProcessor
	database      db.DatabaseConnection
	echo          *echo.Echo
	port          int
}

// Config holds server configuration
type Config struct {
	Port        int
	Verbose     bool
	UseMemgraph bool
}

// NewServer creates a new server instance
func NewServer(config Config) (*Server, error) {
	// Initialize components
	textProcessor := core.NewTextProcessor()
	codeProcessor := core.NewCodeProcessor()

	var database db.DatabaseConnection
	if config.UseMemgraph {
		memgraphDB := db.NewMemgraphDatabase("bolt://localhost:7687", "", "")
		if err := memgraphDB.Connect(); err != nil {
			return nil, fmt.Errorf("failed to connect to Memgraph: %w", err)
		}
		database = memgraphDB
	} else {
		database = db.NewInMemoryDatabase()
		if err := database.Connect(); err != nil {
			return nil, fmt.Errorf("failed to connect to in-memory database: %w", err)
		}
	}

	generator := core.NewKnowledgeGraphGenerator(textProcessor, database)

	// Create Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Hide Echo banner if not verbose
	if !config.Verbose {
		e.HideBanner = true
	}

	server := &Server{
		generator:     generator,
		codeProcessor: codeProcessor,
		database:      database,
		echo:          e,
		port:          config.Port,
	}

	server.setupRoutes()

	return server, nil
}

// setupRoutes configures all the API routes
func (s *Server) setupRoutes() {
	// API group
	api := s.echo.Group("/api")

	// Analysis endpoints
	api.POST("/analyze/text", s.analyzeTextHandler())
	api.POST("/analyze/file", s.analyzeFileHandler())
	api.POST("/analyze/codebase", s.analyzeCodebaseHandler())

	// Query endpoints
	api.GET("/stats", s.getStatsHandler())
	api.GET("/entities", s.getEntitiesHandler())
	api.GET("/relationships", s.getRelationshipsHandler())
	api.GET("/query", s.queryHandler())

	// Health check
	s.echo.GET("/health", s.healthHandler())

	// API documentation endpoint
	s.echo.GET("/", s.docsHandler())
}

// Start starts the server
func (s *Server) Start() error {
	return s.echo.Start(fmt.Sprintf(":%d", s.port))
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	if memgraphDB, ok := s.database.(*db.MemgraphDatabase); ok {
		memgraphDB.Disconnect()
	}
	return nil
}

// Request/Response types
type AnalyzeTextRequest struct {
	Text string `json:"text" validate:"required"`
}

type AnalyzeFileRequest struct {
	FilePath string `json:"filePath" validate:"required"`
}

type AnalyzeCodebaseRequest struct {
	Directory string `json:"directory" validate:"required"`
}

type AnalysisResponse struct {
	Success       bool                   `json:"success"`
	Message       string                 `json:"message,omitempty"`
	Entities      []graph.Entity         `json:"entities,omitempty"`
	Relationships []graph.Relationship   `json:"relationships,omitempty"`
	Statistics    *graph.GraphStatistics `json:"statistics,omitempty"`
}

type APIDocsResponse struct {
	Service   string                `json:"service"`
	Version   string                `json:"version"`
	Endpoints []EndpointDoc         `json:"endpoints"`
	Examples  map[string]ExampleDoc `json:"examples"`
}

type EndpointDoc struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Description string `json:"description"`
}

type ExampleDoc struct {
	Description string      `json:"description"`
	Request     interface{} `json:"request,omitempty"`
	Response    interface{} `json:"response"`
}

// Handler methods
func (s *Server) analyzeTextHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		var req AnalyzeTextRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, AnalysisResponse{
				Success: false,
				Message: "Invalid request format",
			})
		}

		if req.Text == "" {
			return c.JSON(http.StatusBadRequest, AnalysisResponse{
				Success: false,
				Message: "Text field is required",
			})
		}

		entities, relationships, err := s.analyzeText(req.Text)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, AnalysisResponse{
				Success: false,
				Message: fmt.Sprintf("Analysis failed: %v", err),
			})
		}

		return c.JSON(http.StatusOK, AnalysisResponse{
			Success:       true,
			Entities:      entities,
			Relationships: relationships,
		})
	}
}

func (s *Server) analyzeFileHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		var req AnalyzeFileRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, AnalysisResponse{
				Success: false,
				Message: "Invalid request format",
			})
		}

		if req.FilePath == "" {
			return c.JSON(http.StatusBadRequest, AnalysisResponse{
				Success: false,
				Message: "FilePath field is required",
			})
		}

		kg, err := s.analyzeFile(req.FilePath)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, AnalysisResponse{
				Success: false,
				Message: fmt.Sprintf("File analysis failed: %v", err),
			})
		}

		// Store in database
		err = s.generator.StoreKnowledgeGraph(kg.Entities, kg.Relationships)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, AnalysisResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to store results: %v", err),
			})
		}

		return c.JSON(http.StatusOK, AnalysisResponse{
			Success:       true,
			Entities:      kg.Entities,
			Relationships: kg.Relationships,
		})
	}
}

func (s *Server) analyzeCodebaseHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		var req AnalyzeCodebaseRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, AnalysisResponse{
				Success: false,
				Message: "Invalid request format",
			})
		}

		if req.Directory == "" {
			return c.JSON(http.StatusBadRequest, AnalysisResponse{
				Success: false,
				Message: "Directory field is required",
			})
		}

		kg, err := s.analyzeCodebase(req.Directory)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, AnalysisResponse{
				Success: false,
				Message: fmt.Sprintf("Codebase analysis failed: %v", err),
			})
		}

		// Store in database
		err = s.generator.StoreKnowledgeGraph(kg.Entities, kg.Relationships)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, AnalysisResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to store results: %v", err),
			})
		}

		return c.JSON(http.StatusOK, AnalysisResponse{
			Success:       true,
			Entities:      kg.Entities,
			Relationships: kg.Relationships,
		})
	}
}

func (s *Server) getStatsHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		stats, err := s.generator.GetGraphStatistics()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, AnalysisResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to get statistics: %v", err),
			})
		}

		return c.JSON(http.StatusOK, AnalysisResponse{
			Success:    true,
			Statistics: stats,
		})
	}
}

func (s *Server) getEntitiesHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		results, err := s.database.Query("MATCH (n) RETURN n", nil)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, AnalysisResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to get entities: %v", err),
			})
		}

		entities := make([]graph.Entity, 0)
		for _, result := range results {
			if entity, ok := result["n"].(graph.Entity); ok {
				entities = append(entities, entity)
			}
		}

		return c.JSON(http.StatusOK, AnalysisResponse{
			Success:  true,
			Entities: entities,
		})
	}
}

func (s *Server) getRelationshipsHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		results, err := s.database.Query("MATCH (a)-[r]->(b) RETURN r", nil)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, AnalysisResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to get relationships: %v", err),
			})
		}

		relationships := make([]graph.Relationship, 0)
		for _, result := range results {
			if rel, ok := result["r"].(graph.Relationship); ok {
				relationships = append(relationships, rel)
			}
		}

		return c.JSON(http.StatusOK, AnalysisResponse{
			Success:       true,
			Relationships: relationships,
		})
	}
}

func (s *Server) queryHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		query := c.QueryParam("q")
		if query == "" {
			return c.JSON(http.StatusBadRequest, AnalysisResponse{
				Success: false,
				Message: "Query parameter 'q' is required",
			})
		}

		results, err := s.generator.QueryKnowledgeGraph(query, nil)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, AnalysisResponse{
				Success: false,
				Message: fmt.Sprintf("Query failed: %v", err),
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
			"results": results,
		})
	}
}

func (s *Server) healthHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		_, isMemgraph := s.database.(*db.MemgraphDatabase)
		return c.JSON(http.StatusOK, map[string]string{
			"status": "healthy",
			"database": func() string {
				if isMemgraph {
					return "memgraph"
				}
				return "in-memory"
			}(),
		})
	}
}

func (s *Server) docsHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		docs := APIDocsResponse{
			Service: "CodeGraphGen API",
			Version: "1.0.0",
			Endpoints: []EndpointDoc{
				{Method: "GET", Path: "/health", Description: "Health check endpoint"},
				{Method: "POST", Path: "/api/analyze/text", Description: "Analyze text content"},
				{Method: "POST", Path: "/api/analyze/file", Description: "Analyze a file"},
				{Method: "POST", Path: "/api/analyze/codebase", Description: "Analyze a codebase directory"},
				{Method: "GET", Path: "/api/stats", Description: "Get knowledge graph statistics"},
				{Method: "GET", Path: "/api/entities", Description: "Get all entities"},
				{Method: "GET", Path: "/api/relationships", Description: "Get all relationships"},
				{Method: "GET", Path: "/api/query", Description: "Execute a query against the graph"},
			},
			Examples: map[string]ExampleDoc{
				"analyze_text": {
					Description: "Analyze a text snippet",
					Request:     AnalyzeTextRequest{Text: "function hello() { return 'world'; }"},
					Response:    AnalysisResponse{Success: true},
				},
				"health_check": {
					Description: "Check server health",
					Response:    map[string]string{"status": "healthy", "database": "in-memory"},
				},
			},
		}

		return c.JSON(http.StatusOK, docs)
	}
}

// Helper methods for analysis
func (s *Server) analyzeText(text string) ([]graph.Entity, []graph.Relationship, error) {
	kg, err := s.generator.GenerateKnowledgeGraph(text)
	if err != nil {
		return nil, nil, err
	}
	return kg.Entities, kg.Relationships, nil
}

func (s *Server) analyzeFile(filePath string) (*graph.KnowledgeGraph, error) {
	entities, relationships, err := s.codeProcessor.AnalyzeCodebase(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to process file: %w", err)
	}

	return &graph.KnowledgeGraph{
		Entities:      entities,
		Relationships: relationships,
	}, nil
}

func (s *Server) analyzeCodebase(directory string) (*graph.KnowledgeGraph, error) {
	entities, relationships, err := s.codeProcessor.AnalyzeCodebase(directory)
	if err != nil {
		return nil, fmt.Errorf("failed to process directory: %w", err)
	}

	return &graph.KnowledgeGraph{
		Entities:      entities,
		Relationships: relationships,
	}, nil
}
