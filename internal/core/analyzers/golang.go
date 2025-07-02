package analyzers

import (
	"path/filepath"
	"regexp"
	"strings"

	"codegraphgen/internal/core/graph"
)

// GoImport represents a Go import statement
type GoImport struct {
	Name       string
	Path       string
	Alias      string
	LineNumber int
}

// GoStruct represents a Go struct
type GoStruct struct {
	Name       string
	LineNumber int
	IsExported bool
	Fields     []GoField
}

// GoField represents a struct field
type GoField struct {
	Name       string
	Type       string
	LineNumber int
	IsExported bool
}

// GoFunction represents a Go function
type GoFunction struct {
	Name        string
	LineNumber  int
	IsExported  bool
	Receiver    string
	Parameters  []string
	ReturnTypes []string
}

// GoInterface represents a Go interface
type GoInterface struct {
	Name       string
	LineNumber int
	IsExported bool
	Methods    []string
}

// GoType represents a Go type definition
type GoType struct {
	Name       string
	LineNumber int
	IsExported bool
	Definition string
}

// GoConstant represents a Go constant
type GoConstant struct {
	Name       string
	LineNumber int
	IsExported bool
	Type       string
	Value      string
}

// FunctionCall represents a function call relationship
type FunctionCall struct {
	Caller     string
	Callee     string
	LineNumber int
}

// GoAnalyzer implements the LanguageAnalyzer interface for Go language
type GoAnalyzer struct{}

// Name returns the name of this analyzer
func (ga *GoAnalyzer) Name() string {
	return "Go Analyzer"
}

// SupportedLanguages returns the languages this analyzer supports
func (ga *GoAnalyzer) SupportedLanguages() []string {
	return []string{"go"}
}

// Analyze analyzes a Go source file
func (ga *GoAnalyzer) Analyze(file graph.CodeFile, fileEntity graph.Entity) ([]graph.Entity, []graph.Relationship, error) {
	return analyzeGoFile(file, fileEntity)
}

// analyzeGoFile analyzes a Go source file
func analyzeGoFile(file graph.CodeFile, fileEntity graph.Entity) ([]graph.Entity, []graph.Relationship, error) {
	entities := []graph.Entity{fileEntity}
	var relationships []graph.Relationship

	content := file.Content

	// Extract package declaration
	packageRegex := regexp.MustCompile(`package\s+(\w+)`)
	if match := packageRegex.FindStringSubmatch(content); len(match) > 1 {
		packageEntity := graph.CreateEntity(match[1], graph.EntityTypePackage, graph.Properties{
			"sourceFile": file.Path,
			"language":   "go",
		})
		entities = append(entities, packageEntity)
		relationships = append(relationships, graph.CreateRelationship(
			fileEntity.ID, packageEntity.ID, graph.RelationshipTypeDefines, nil))
	}

	// Extract imports
	imports := extractGoImports(content)
	for _, imp := range imports {
		importEntity := graph.CreateEntity(imp.Name, graph.EntityTypeImport, graph.Properties{
			"source":     imp.Path,
			"alias":      imp.Alias,
			"lineNumber": imp.LineNumber,
			"language":   "go",
		})
		entities = append(entities, importEntity)
		relationships = append(relationships, graph.CreateRelationship(
			fileEntity.ID, importEntity.ID, graph.RelationshipTypeImports, nil))
	}

	// Extract structs (similar to classes)
	structs := extractGoStructs(content)
	for _, st := range structs {
		structEntity := graph.CreateEntity(st.Name, graph.EntityTypeClass, graph.Properties{
			"sourceFile": file.Path,
			"lineNumber": st.LineNumber,
			"isExported": st.IsExported,
			"language":   "go",
			"structType": true,
		})
		entities = append(entities, structEntity)
		relationships = append(relationships, graph.CreateRelationship(
			fileEntity.ID, structEntity.ID, graph.RelationshipTypeDefines, nil))

		// Extract struct fields
		for _, field := range st.Fields {
			fieldEntity := graph.CreateEntity(field.Name, graph.EntityTypeProperty, graph.Properties{
				"sourceFile": file.Path,
				"lineNumber": field.LineNumber,
				"type":       field.Type,
				"isExported": field.IsExported,
				"language":   "go",
			})
			entities = append(entities, fieldEntity)
			relationships = append(relationships, graph.CreateRelationship(
				structEntity.ID, fieldEntity.ID, graph.RelationshipTypeContains, nil))
		}
	}

	// Extract functions
	functions := extractGoFunctions(content)
	for _, fn := range functions {
		funcEntity := graph.CreateEntity(fn.Name, graph.EntityTypeFunction, graph.Properties{
			"sourceFile":  file.Path,
			"lineNumber":  fn.LineNumber,
			"isExported":  fn.IsExported,
			"receiver":    fn.Receiver,
			"parameters":  fn.Parameters,
			"returnTypes": fn.ReturnTypes,
			"language":    "go",
		})
		entities = append(entities, funcEntity)

		if fn.Receiver != "" {
			// This is a method - find the receiver struct
			// Extract the receiver type name from syntax like "db *MemgraphDatabase" or "m MemgraphDatabase"
			receiverType := extractReceiverType(fn.Receiver)

			// Look for the receiver struct in entities
			var foundReceiver bool
			for _, entity := range entities {
				if entity.Type == graph.EntityTypeClass && entity.Label == receiverType {
					relationships = append(relationships, graph.CreateRelationship(
						entity.ID, funcEntity.ID, graph.RelationshipTypeContains, nil))
					foundReceiver = true
					break
				}
			}

			// If we didn't find the receiver struct in current entities, still connect to file
			if !foundReceiver {
				relationships = append(relationships, graph.CreateRelationship(
					fileEntity.ID, funcEntity.ID, graph.RelationshipTypeDefines, nil))
			}
		} else {
			// This is a standalone function
			relationships = append(relationships, graph.CreateRelationship(
				fileEntity.ID, funcEntity.ID, graph.RelationshipTypeDefines, nil))
		}
	}

	// Extract function calls and create CALLS relationships
	functionCalls := extractFunctionCalls(content, functions)
	for _, call := range functionCalls {
		// Find the calling and called function entities
		var callerEntity, calleeEntity *graph.Entity
		for i := range entities {
			entity := &entities[i]
			if entity.Type == graph.EntityTypeFunction {
				if entity.Label == call.Caller {
					callerEntity = entity
				}
				if entity.Label == call.Callee {
					calleeEntity = entity
				}
			}
		}

		// Create CALLS relationship if both entities found
		if callerEntity != nil && calleeEntity != nil && callerEntity.ID != calleeEntity.ID {
			relationships = append(relationships, graph.CreateRelationship(
				callerEntity.ID, calleeEntity.ID, graph.RelationshipTypeCalls, graph.Properties{
					"lineNumber": call.LineNumber,
				}))
		}
	}

	// Extract interfaces
	interfaces := extractGoInterfaces(content)
	for _, iface := range interfaces {
		interfaceEntity := graph.CreateEntity(iface.Name, graph.EntityTypeInterface, graph.Properties{
			"sourceFile": file.Path,
			"lineNumber": iface.LineNumber,
			"isExported": iface.IsExported,
			"methods":    iface.Methods,
			"language":   "go",
		})
		entities = append(entities, interfaceEntity)
		relationships = append(relationships, graph.CreateRelationship(
			fileEntity.ID, interfaceEntity.ID, graph.RelationshipTypeDefines, nil))
	}

	// Extract type definitions
	types := extractGoTypes(content)
	for _, typ := range types {
		typeEntity := graph.CreateEntity(typ.Name, graph.EntityTypeType, graph.Properties{
			"sourceFile": file.Path,
			"lineNumber": typ.LineNumber,
			"isExported": typ.IsExported,
			"definition": typ.Definition,
			"language":   "go",
		})
		entities = append(entities, typeEntity)
		relationships = append(relationships, graph.CreateRelationship(
			fileEntity.ID, typeEntity.ID, graph.RelationshipTypeDefines, nil))
	}

	// Extract constants
	constants := extractGoConstants(content)
	for _, constant := range constants {
		constEntity := graph.CreateEntity(constant.Name, graph.EntityTypeConstant, graph.Properties{
			"sourceFile": file.Path,
			"lineNumber": constant.LineNumber,
			"isExported": constant.IsExported,
			"type":       constant.Type,
			"value":      constant.Value,
			"language":   "go",
		})
		entities = append(entities, constEntity)
		relationships = append(relationships, graph.CreateRelationship(
			fileEntity.ID, constEntity.ID, graph.RelationshipTypeDefines, nil))
	}

	return entities, relationships, nil
}

func extractGoImports(content string) []GoImport {
	var imports []GoImport
	lines := strings.Split(content, "\n")

	// Handle single import
	singleImportRegex := regexp.MustCompile(`import\s+"([^"]+)"`)
	// Handle aliased import
	aliasImportRegex := regexp.MustCompile(`import\s+(\w+)\s+"([^"]+)"`)
	// Handle import block
	importBlockRegex := regexp.MustCompile(`import\s*\(`)

	inImportBlock := false

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Check for import block start
		if importBlockRegex.MatchString(line) {
			inImportBlock = true
			continue
		}

		// Check for import block end
		if inImportBlock && strings.Contains(line, ")") {
			inImportBlock = false
			continue
		}

		// Process imports within block
		if inImportBlock {
			if match := regexp.MustCompile(`"([^"]+)"`).FindStringSubmatch(line); len(match) > 1 {
				importPath := match[1]
				name := filepath.Base(importPath)
				imports = append(imports, GoImport{
					Name:       name,
					Path:       importPath,
					LineNumber: i + 1,
				})
			}
			continue
		}

		// Process single line imports
		if match := aliasImportRegex.FindStringSubmatch(line); len(match) > 2 {
			imports = append(imports, GoImport{
				Name:       match[1],
				Path:       match[2],
				Alias:      match[1],
				LineNumber: i + 1,
			})
		} else if match := singleImportRegex.FindStringSubmatch(line); len(match) > 1 {
			importPath := match[1]
			name := filepath.Base(importPath)
			imports = append(imports, GoImport{
				Name:       name,
				Path:       importPath,
				LineNumber: i + 1,
			})
		}
	}

	return imports
}

func extractGoStructs(content string) []GoStruct {
	var structs []GoStruct
	lines := strings.Split(content, "\n")

	structRegex := regexp.MustCompile(`type\s+(\w+)\s+struct`)

	for i, line := range lines {
		if match := structRegex.FindStringSubmatch(line); len(match) > 1 {
			structName := match[1]
			isExported := len(structName) > 0 && structName[0] >= 'A' && structName[0] <= 'Z'

			// Extract fields (simplified)
			var fields []GoField
			// In a real implementation, you'd parse the struct body

			structs = append(structs, GoStruct{
				Name:       structName,
				LineNumber: i + 1,
				IsExported: isExported,
				Fields:     fields,
			})
		}
	}

	return structs
}

func extractGoFunctions(content string) []GoFunction {
	var functions []GoFunction
	lines := strings.Split(content, "\n")

	// Function regex that handles receivers
	funcRegex := regexp.MustCompile(`func\s*(?:\([^)]*\))?\s*(\w+)\s*\([^)]*\)`)
	receiverRegex := regexp.MustCompile(`func\s*\(([^)]*)\)\s*(\w+)`)

	for i, line := range lines {
		line = strings.TrimSpace(line)

		if match := funcRegex.FindStringSubmatch(line); len(match) > 1 {
			funcName := match[1]
			isExported := len(funcName) > 0 && funcName[0] >= 'A' && funcName[0] <= 'Z'

			var receiver string
			if receiverMatch := receiverRegex.FindStringSubmatch(line); len(receiverMatch) > 2 {
				receiver = strings.TrimSpace(receiverMatch[1])
				funcName = receiverMatch[2]
			}

			functions = append(functions, GoFunction{
				Name:       funcName,
				LineNumber: i + 1,
				IsExported: isExported,
				Receiver:   receiver,
			})
		}
	}

	return functions
}

func extractGoInterfaces(content string) []GoInterface {
	var interfaces []GoInterface
	lines := strings.Split(content, "\n")

	interfaceRegex := regexp.MustCompile(`type\s+(\w+)\s+interface`)

	for i, line := range lines {
		if match := interfaceRegex.FindStringSubmatch(line); len(match) > 1 {
			interfaceName := match[1]
			isExported := len(interfaceName) > 0 && interfaceName[0] >= 'A' && interfaceName[0] <= 'Z'

			interfaces = append(interfaces, GoInterface{
				Name:       interfaceName,
				LineNumber: i + 1,
				IsExported: isExported,
			})
		}
	}

	return interfaces
}

func extractGoTypes(content string) []GoType {
	var types []GoType
	lines := strings.Split(content, "\n")

	typeRegex := regexp.MustCompile(`type\s+(\w+)\s+(.+)`)

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if match := typeRegex.FindStringSubmatch(line); len(match) > 2 {
			typeName := match[1]
			definition := match[2]

			// Skip struct and interface definitions (handled separately)
			if strings.Contains(definition, "struct") || strings.Contains(definition, "interface") {
				continue
			}

			isExported := len(typeName) > 0 && typeName[0] >= 'A' && typeName[0] <= 'Z'

			types = append(types, GoType{
				Name:       typeName,
				LineNumber: i + 1,
				IsExported: isExported,
				Definition: definition,
			})
		}
	}

	return types
}

func extractGoConstants(content string) []GoConstant {
	var constants []GoConstant
	lines := strings.Split(content, "\n")

	constRegex := regexp.MustCompile(`const\s+(\w+)(?:\s+(\w+))?\s*=\s*(.+)`)

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if match := constRegex.FindStringSubmatch(line); len(match) > 3 {
			constName := match[1]
			constType := match[2]
			constValue := match[3]

			isExported := len(constName) > 0 && constName[0] >= 'A' && constName[0] <= 'Z'

			constants = append(constants, GoConstant{
				Name:       constName,
				LineNumber: i + 1,
				IsExported: isExported,
				Type:       constType,
				Value:      constValue,
			})
		}
	}

	return constants
}

// extractFunctionCalls extracts function calls from Go code
func extractFunctionCalls(content string, functions []GoFunction) []FunctionCall {
	var calls []FunctionCall
	lines := strings.Split(content, "\n")

	// Create a map of function names for quick lookup
	functionNames := make(map[string]bool)
	for _, fn := range functions {
		functionNames[fn.Name] = true
	}

	// Track which function we're currently inside
	var currentFunction string

	// Function call regex patterns
	directCallRegex := regexp.MustCompile(`(\w+)\s*\(`)          // functionName(
	methodCallRegex := regexp.MustCompile(`\.(\w+)\s*\(`)        // .methodName(
	receiverCallRegex := regexp.MustCompile(`(\w+)\.(\w+)\s*\(`) // receiver.method(

	for i, line := range lines {
		line = strings.TrimSpace(line)
		lineNumber := i + 1

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") {
			continue
		}

		// Check if we're entering a new function
		if match := regexp.MustCompile(`func\s*(?:\([^)]*\))?\s*(\w+)\s*\(`).FindStringSubmatch(line); len(match) > 1 {
			currentFunction = match[1]
			continue
		}

		if currentFunction == "" {
			continue
		}

		// Find direct function calls (functionName())
		if matches := directCallRegex.FindAllStringSubmatch(line, -1); len(matches) > 0 {
			for _, match := range matches {
				if len(match) > 1 {
					calledFunc := match[1]
					// Only count calls to functions we've identified
					if functionNames[calledFunc] && calledFunc != currentFunction {
						calls = append(calls, FunctionCall{
							Caller:     currentFunction,
							Callee:     calledFunc,
							LineNumber: lineNumber,
						})
					}
				}
			}
		}

		// Find method calls (.methodName())
		if matches := methodCallRegex.FindAllStringSubmatch(line, -1); len(matches) > 0 {
			for _, match := range matches {
				if len(match) > 1 {
					calledMethod := match[1]
					// Only count calls to methods we've identified
					if functionNames[calledMethod] && calledMethod != currentFunction {
						calls = append(calls, FunctionCall{
							Caller:     currentFunction,
							Callee:     calledMethod,
							LineNumber: lineNumber,
						})
					}
				}
			}
		}

		// Find receiver.method calls
		if matches := receiverCallRegex.FindAllStringSubmatch(line, -1); len(matches) > 0 {
			for _, match := range matches {
				if len(match) > 2 {
					calledMethod := match[2]
					// Only count calls to methods we've identified
					if functionNames[calledMethod] && calledMethod != currentFunction {
						calls = append(calls, FunctionCall{
							Caller:     currentFunction,
							Callee:     calledMethod,
							LineNumber: lineNumber,
						})
					}
				}
			}
		}
	}

	return calls
}

// extractReceiverType extracts the type name from a Go receiver string
// e.g., "db *MemgraphDatabase" -> "MemgraphDatabase"
// e.g., "m MemgraphDatabase" -> "MemgraphDatabase"
func extractReceiverType(receiver string) string {
	// Remove whitespace and split by spaces
	parts := strings.Fields(strings.TrimSpace(receiver))
	if len(parts) == 0 {
		return ""
	}

	// Get the last part (type name)
	typeName := parts[len(parts)-1]

	// Remove pointer indicator if present
	typeName = strings.TrimPrefix(typeName, "*")

	// Remove any package prefix if present (e.g., "pkg.Type" -> "Type")
	if dotIndex := strings.LastIndex(typeName, "."); dotIndex != -1 {
		typeName = typeName[dotIndex+1:]
	}

	return typeName
}
