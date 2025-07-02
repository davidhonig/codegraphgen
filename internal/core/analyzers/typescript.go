package analyzers

import (
	"codegraphgen/internal/core/graph"
	"regexp"
	"strings"
)

// TypeScript-specific data structures
type TypeScriptImport struct {
	Name        string
	Source      string
	IsDefault   bool
	IsNamespace bool
	LineNumber  int
}

type TypeScriptClass struct {
	Name       string
	LineNumber int
	IsAbstract bool
	IsExported bool
	Extends    []string
	Implements []string
	Methods    []TypeScriptMethod
	Properties []TypeScriptProperty
}

type TypeScriptMethod struct {
	Name       string
	LineNumber int
	Visibility string
	IsStatic   bool
	IsAsync    bool
	Parameters []string
	ReturnType string
}

type TypeScriptProperty struct {
	Name       string
	LineNumber int
	Visibility string
	IsStatic   bool
	Type       string
	IsReadonly bool
}

type TypeScriptFunction struct {
	Name       string
	LineNumber int
	IsAsync    bool
	IsExported bool
	Parameters []string
	ReturnType string
}

type TypeScriptInterface struct {
	Name       string
	LineNumber int
	IsExported bool
	Extends    []string
}

type TypeScriptType struct {
	Name       string
	LineNumber int
	IsExported bool
	Definition string
}

// TypeScriptAnalyzer implements the LanguageAnalyzer interface for TypeScript/JavaScript

type TypeScriptAnalyzer struct{}

func (tsa *TypeScriptAnalyzer) Name() string { return "TypeScript Analyzer" }
func (tsa *TypeScriptAnalyzer) SupportedLanguages() []string {
	return []string{"typescript", "javascript"}
}
func (tsa *TypeScriptAnalyzer) Analyze(file graph.CodeFile, fileEntity graph.Entity) ([]graph.Entity, []graph.Relationship, error) {
	return analyzeTypeScriptFile(file, fileEntity)
}

// analyzeTypeScriptFile analyzes a TypeScript/JavaScript file

func analyzeTypeScriptFile(file graph.CodeFile, fileEntity graph.Entity) ([]graph.Entity, []graph.Relationship, error) {
	entities := []graph.Entity{fileEntity}
	var relationships []graph.Relationship

	content := file.Content

	// Extract imports
	imports := extractTypeScriptImports(content)
	for _, imp := range imports {
		importEntity := graph.CreateEntity(imp.Name, graph.EntityTypeImport, graph.Properties{
			"source":      imp.Source,
			"isDefault":   imp.IsDefault,
			"isNamespace": imp.IsNamespace,
			"lineNumber":  imp.LineNumber,
			"language":    file.Language,
		})
		entities = append(entities, importEntity)
		relationships = append(relationships, graph.CreateRelationship(
			fileEntity.ID, importEntity.ID, graph.RelationshipTypeImports, nil))
	}

	// Extract classes
	classes := extractTypeScriptClasses(content)
	for _, cls := range classes {
		classEntity := graph.CreateEntity(cls.Name, graph.EntityTypeClass, graph.Properties{
			"sourceFile": file.Path,
			"lineNumber": cls.LineNumber,
			"isAbstract": cls.IsAbstract,
			"isExported": cls.IsExported,
			"extends":    cls.Extends,
			"implements": cls.Implements,
			"language":   file.Language,
		})
		entities = append(entities, classEntity)
		relationships = append(relationships, graph.CreateRelationship(
			fileEntity.ID, classEntity.ID, graph.RelationshipTypeDefines, nil))

		// Extract methods
		for _, method := range cls.Methods {
			methodEntity := graph.CreateEntity(method.Name, graph.EntityTypeMethod, graph.Properties{
				"sourceFile": file.Path,
				"lineNumber": method.LineNumber,
				"visibility": method.Visibility,
				"isStatic":   method.IsStatic,
				"isAsync":    method.IsAsync,
				"parameters": method.Parameters,
				"returnType": method.ReturnType,
				"language":   file.Language,
			})
			entities = append(entities, methodEntity)
			relationships = append(relationships, graph.CreateRelationship(
				classEntity.ID, methodEntity.ID, graph.RelationshipTypeContains, nil))
		}

		// Extract properties
		for _, prop := range cls.Properties {
			propEntity := graph.CreateEntity(prop.Name, graph.EntityTypeProperty, graph.Properties{
				"sourceFile": file.Path,
				"lineNumber": prop.LineNumber,
				"visibility": prop.Visibility,
				"isStatic":   prop.IsStatic,
				"type":       prop.Type,
				"isReadonly": prop.IsReadonly,
				"language":   file.Language,
			})
			entities = append(entities, propEntity)
			relationships = append(relationships, graph.CreateRelationship(
				classEntity.ID, propEntity.ID, graph.RelationshipTypeContains, nil))
		}
	}

	// Extract functions
	functions := extractTypeScriptFunctions(content)
	for _, fn := range functions {
		funcEntity := graph.CreateEntity(fn.Name, graph.EntityTypeFunction, graph.Properties{
			"sourceFile": file.Path,
			"lineNumber": fn.LineNumber,
			"isAsync":    fn.IsAsync,
			"isExported": fn.IsExported,
			"parameters": fn.Parameters,
			"returnType": fn.ReturnType,
			"language":   file.Language,
		})
		entities = append(entities, funcEntity)
		relationships = append(relationships, graph.CreateRelationship(
			fileEntity.ID, funcEntity.ID, graph.RelationshipTypeDefines, nil))
	}

	// Extract interfaces
	interfaces := extractTypeScriptInterfaces(content)
	for _, iface := range interfaces {
		ifaceEntity := graph.CreateEntity(iface.Name, graph.EntityTypeInterface, graph.Properties{
			"sourceFile": file.Path,
			"lineNumber": iface.LineNumber,
			"isExported": iface.IsExported,
			"extends":    iface.Extends,
			"language":   file.Language,
		})
		entities = append(entities, ifaceEntity)
		relationships = append(relationships, graph.CreateRelationship(
			fileEntity.ID, ifaceEntity.ID, graph.RelationshipTypeDefines, nil))
	}

	// Extract types
	types := extractTypeScriptTypes(content)
	for _, typ := range types {
		typeEntity := graph.CreateEntity(typ.Name, graph.EntityTypeType, graph.Properties{
			"sourceFile": file.Path,
			"lineNumber": typ.LineNumber,
			"isExported": typ.IsExported,
			"definition": typ.Definition,
			"language":   file.Language,
		})
		entities = append(entities, typeEntity)
		relationships = append(relationships, graph.CreateRelationship(
			fileEntity.ID, typeEntity.ID, graph.RelationshipTypeDefines, nil))
	}

	return entities, relationships, nil
}

// TypeScript extraction methods
func extractTypeScriptImports(content string) []TypeScriptImport {
	var imports []TypeScriptImport
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// ES6 imports
		importRegex := regexp.MustCompile(`import\s+(.+?)\s+from\s+['"](.+?)['"]`)
		if match := importRegex.FindStringSubmatch(line); len(match) > 2 {
			importClause := match[1]
			source := match[2]

			// Handle different import types
			if strings.Contains(importClause, "{") {
				// Named imports
				namedRegex := regexp.MustCompile(`\{(.+?)\}`)
				if namedMatch := namedRegex.FindStringSubmatch(importClause); len(namedMatch) > 1 {
					namedImports := namedMatch[1]
					names := strings.Split(namedImports, ",")
					for _, name := range names {
						name = strings.TrimSpace(name)
						if strings.Contains(name, " as ") {
							name = strings.Split(name, " as ")[0]
						}
						name = strings.TrimSpace(name)
						imports = append(imports, TypeScriptImport{
							Name:        name,
							Source:      source,
							IsDefault:   false,
							IsNamespace: false,
							LineNumber:  i + 1,
						})
					}
				}
			} else if strings.Contains(importClause, "* as ") {
				// Namespace import
				namespaceName := strings.TrimSpace(strings.Replace(importClause, "* as ", "", 1))
				imports = append(imports, TypeScriptImport{
					Name:        namespaceName,
					Source:      source,
					IsDefault:   false,
					IsNamespace: true,
					LineNumber:  i + 1,
				})
			} else {
				// Default import
				imports = append(imports, TypeScriptImport{
					Name:        strings.TrimSpace(importClause),
					Source:      source,
					IsDefault:   true,
					IsNamespace: false,
					LineNumber:  i + 1,
				})
			}
		}
	}

	return imports
}

func extractTypeScriptClasses(content string) []TypeScriptClass {
	var classes []TypeScriptClass
	lines := strings.Split(content, "\n")

	classRegex := regexp.MustCompile(`(?:export\s+)?(?:abstract\s+)?class\s+(\w+)(?:\s+extends\s+(\w+))?(?:\s+implements\s+(.+?))?`)

	for i, line := range lines {
		line = strings.TrimSpace(line)

		if match := classRegex.FindStringSubmatch(line); len(match) > 1 {
			className := match[1]
			extendsClause := match[2]
			implementsClause := match[3]

			var extends []string
			if extendsClause != "" {
				extends = []string{extendsClause}
			}

			var implements []string
			if implementsClause != "" {
				implements = strings.Split(implementsClause, ",")
				for j, impl := range implements {
					implements[j] = strings.TrimSpace(impl)
				}
			}

			classInfo := TypeScriptClass{
				Name:       className,
				LineNumber: i + 1,
				IsAbstract: strings.Contains(line, "abstract"),
				IsExported: strings.Contains(line, "export"),
				Extends:    extends,
				Implements: implements,
				Methods:    []TypeScriptMethod{},
				Properties: []TypeScriptProperty{},
			}

			// Extract methods and properties from class body (simplified)
			// In a full implementation, you'd parse the class body more thoroughly
			classes = append(classes, classInfo)
		}
	}

	return classes
}

func extractTypeScriptFunctions(content string) []TypeScriptFunction {
	var functions []TypeScriptFunction
	lines := strings.Split(content, "\n")

	// Function declaration
	funcRegex := regexp.MustCompile(`(?:export\s+)?(?:async\s+)?function\s+(\w+)\s*\(`)
	// Arrow function
	arrowRegex := regexp.MustCompile(`(?:export\s+)?(?:const|let|var)\s+(\w+)\s*=\s*(?:async\s+)?\(`)

	for i, line := range lines {
		line = strings.TrimSpace(line)

		if match := funcRegex.FindStringSubmatch(line); len(match) > 1 {
			functions = append(functions, TypeScriptFunction{
				Name:       match[1],
				LineNumber: i + 1,
				IsAsync:    strings.Contains(line, "async"),
				IsExported: strings.Contains(line, "export"),
				Parameters: []string{}, // Simplified for now
				ReturnType: "unknown",
			})
		} else if match := arrowRegex.FindStringSubmatch(line); len(match) > 1 {
			functions = append(functions, TypeScriptFunction{
				Name:       match[1],
				LineNumber: i + 1,
				IsAsync:    strings.Contains(line, "async"),
				IsExported: strings.Contains(line, "export"),
				Parameters: []string{}, // Simplified for now
				ReturnType: "unknown",
			})
		}
	}

	return functions
}

func extractTypeScriptInterfaces(content string) []TypeScriptInterface {
	var interfaces []TypeScriptInterface
	lines := strings.Split(content, "\n")

	interfaceRegex := regexp.MustCompile(`(?:export\s+)?interface\s+(\w+)(?:\s+extends\s+(.+?))?`)

	for i, line := range lines {
		line = strings.TrimSpace(line)

		if match := interfaceRegex.FindStringSubmatch(line); len(match) > 1 {
			extendsClause := match[2]
			var extends []string
			if extendsClause != "" {
				extends = strings.Split(extendsClause, ",")
				for j, ext := range extends {
					extends[j] = strings.TrimSpace(ext)
				}
			}

			interfaces = append(interfaces, TypeScriptInterface{
				Name:       match[1],
				LineNumber: i + 1,
				IsExported: strings.Contains(line, "export"),
				Extends:    extends,
			})
		}
	}

	return interfaces
}

func extractTypeScriptTypes(content string) []TypeScriptType {
	var types []TypeScriptType
	lines := strings.Split(content, "\n")

	typeRegex := regexp.MustCompile(`(?:export\s+)?type\s+(\w+)\s*=\s*(.+)`)

	for i, line := range lines {
		line = strings.TrimSpace(line)

		if match := typeRegex.FindStringSubmatch(line); len(match) > 2 {
			types = append(types, TypeScriptType{
				Name:       match[1],
				LineNumber: i + 1,
				IsExported: strings.Contains(line, "export"),
				Definition: match[2],
			})
		}
	}

	return types
}
