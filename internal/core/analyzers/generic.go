package analyzers

import "codegraphgen/internal/core/graph"

type GenericAnalyzer struct{}

func (ga *GenericAnalyzer) Name() string                 { return "Generic Analyzer" }
func (ga *GenericAnalyzer) SupportedLanguages() []string { return []string{"unknown"} }
func (ga *GenericAnalyzer) Analyze(file graph.CodeFile, fileEntity graph.Entity) ([]graph.Entity, []graph.Relationship, error) {
	return []graph.Entity{fileEntity}, []graph.Relationship{}, nil
}
