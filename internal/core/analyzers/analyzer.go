package analyzers

import "codegraphgen/internal/core/graph"

type LanguageAnalyzer interface {
	Analyze(file graph.CodeFile, fileEntity graph.Entity) ([]graph.Entity, []graph.Relationship, error)
	SupportedLanguages() []string
	Name() string
}

type AnalyzerRegistry struct {
	analyzers map[string]LanguageAnalyzer
}

func NewAnalyzerRegistry() *AnalyzerRegistry {
	registry := &AnalyzerRegistry{
		analyzers: make(map[string]LanguageAnalyzer),
	}
	registry.RegisterAnalyzer(&GoAnalyzer{})
	registry.RegisterAnalyzer(&TypeScriptAnalyzer{})
	registry.RegisterAnalyzer(&PythonAnalyzer{})
	registry.RegisterAnalyzer(&JavaAnalyzer{})
	registry.RegisterAnalyzer(&JSONAnalyzer{})
	registry.RegisterAnalyzer(&GenericAnalyzer{})
	return registry
}

func (ar *AnalyzerRegistry) RegisterAnalyzer(analyzer LanguageAnalyzer) {
	for _, lang := range analyzer.SupportedLanguages() {
		ar.analyzers[lang] = analyzer
	}
}

func (ar *AnalyzerRegistry) GetAnalyzer(language string) LanguageAnalyzer {
	if analyzer, exists := ar.analyzers[language]; exists {
		return analyzer
	}
	return &GenericAnalyzer{}
}

func (ar *AnalyzerRegistry) ListAnalyzers() map[string]LanguageAnalyzer {
	return ar.analyzers
}
