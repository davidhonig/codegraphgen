package core

import (
	"codegraphgen/internal/core/analyzers"
	"codegraphgen/internal/core/graph"
)

// LanguageAnalyzer defines the interface for language-specific code analyzers
type LanguageAnalyzer interface {
	// Analyze analyzes a code file and returns entities and relationships
	Analyze(file graph.CodeFile, fileEntity graph.Entity) ([]graph.Entity, []graph.Relationship, error)

	// SupportedLanguages returns the list of languages this analyzer supports
	SupportedLanguages() []string

	// Name returns the name of this analyzer
	Name() string
}

// AnalyzerRegistry manages language analyzers
type AnalyzerRegistry struct {
	analyzers map[string]LanguageAnalyzer
}

// NewAnalyzerRegistry creates a new analyzer registry
func NewAnalyzerRegistry() *AnalyzerRegistry {
	registry := &AnalyzerRegistry{
		analyzers: make(map[string]LanguageAnalyzer),
	}

	// Register all available analyzers
	registry.RegisterAnalyzer(&analyzers.GoAnalyzer{})
	registry.RegisterAnalyzer(&analyzers.TypeScriptAnalyzer{})
	registry.RegisterAnalyzer(&analyzers.PythonAnalyzer{})
	registry.RegisterAnalyzer(&analyzers.JavaAnalyzer{})
	registry.RegisterAnalyzer(&analyzers.JSONAnalyzer{})
	registry.RegisterAnalyzer(&analyzers.GenericAnalyzer{})

	return registry
}

// RegisterAnalyzer registers a language analyzer
func (ar *AnalyzerRegistry) RegisterAnalyzer(analyzer LanguageAnalyzer) {
	for _, lang := range analyzer.SupportedLanguages() {
		ar.analyzers[lang] = analyzer
	}
}

// GetAnalyzer returns the analyzer for a specific language
func (ar *AnalyzerRegistry) GetAnalyzer(language string) LanguageAnalyzer {
	if analyzer, exists := ar.analyzers[language]; exists {
		return analyzer
	}
	// Return generic analyzer as fallback
	return &analyzers.GenericAnalyzer{}
}

// ListAnalyzers returns all registered analyzers
func (ar *AnalyzerRegistry) ListAnalyzers() map[string]LanguageAnalyzer {
	return ar.analyzers
}
