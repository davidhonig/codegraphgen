package core

import (
	"codegraphgen/internal/core/graph"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CodeProcessor handles analysis of source code files
type CodeProcessor struct {
	*TextProcessor
	supportedExtensions map[string]bool
	languageMap         map[string]string
	analyzerRegistry    *AnalyzerRegistry
}

// NewCodeProcessor creates a new CodeProcessor instance
func NewCodeProcessor() *CodeProcessor {
	supportedExtensions := map[string]bool{
		".ts":   true,
		".js":   true,
		".tsx":  true,
		".jsx":  true,
		".py":   true,
		".java": true,
		".cpp":  true,
		".c":    true,
		".h":    true,
		".hpp":  true,
		".cs":   true,
		".go":   true,
		".rs":   true,
		".rb":   true,
		".php":  true,
		".json": true,
		".yaml": true,
		".yml":  true,
		".xml":  true,
		".md":   true,
		".txt":  true,
		".sql":  true,
	}

	languageMap := map[string]string{
		".ts":   "typescript",
		".tsx":  "typescript",
		".js":   "javascript",
		".jsx":  "javascript",
		".py":   "python",
		".java": "java",
		".cpp":  "cpp",
		".c":    "c",
		".h":    "c",
		".hpp":  "cpp",
		".cs":   "csharp",
		".go":   "go",
		".rs":   "rust",
		".rb":   "ruby",
		".php":  "php",
		".json": "json",
		".yaml": "yaml",
		".yml":  "yaml",
		".xml":  "xml",
		".md":   "markdown",
		".sql":  "sql",
	}

	return &CodeProcessor{
		TextProcessor:       NewTextProcessor(),
		supportedExtensions: supportedExtensions,
		languageMap:         languageMap,
		analyzerRegistry:    NewAnalyzerRegistry(),
	}
}

// AnalyzeCodebase analyzes an entire codebase directory
func (cp *CodeProcessor) AnalyzeCodebase(rootPath string) ([]graph.Entity, []graph.Relationship, error) {
	fmt.Printf("🔍 Analyzing codebase at: %s\n", rootPath)

	files, err := cp.scanDirectory(rootPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to scan directory: %w", err)
	}

	var allEntities []graph.Entity
	var allRelationships []graph.Relationship

	// Create directory structure entities
	directories := cp.extractDirectories(files)
	for _, dir := range directories {
		allEntities = append(allEntities, cp.createDirectoryEntity(dir, rootPath))
	}

	// Process each file
	for _, file := range files {
		fmt.Printf("📄 Processing: %s\n", file.Path)

		entities, relationships, err := cp.analyzeFile(file)
		if err != nil {
			log.Printf("⚠️ Failed to process %s: %v", file.Path, err)
			continue
		}

		allEntities = append(allEntities, entities...)
		allRelationships = append(allRelationships, relationships...)

		// Create file-to-directory relationships
		fileRelationships := cp.createFileDirectoryRelationships(file, allEntities)
		allRelationships = append(allRelationships, fileRelationships...)
	}

	// Create import/dependency relationships
	importRelationships := cp.createImportRelationships(allEntities)
	allRelationships = append(allRelationships, importRelationships...)

	fmt.Printf("✅ Analyzed %d files, found %d entities and %d relationships\n",
		len(files), len(allEntities), len(allRelationships))

	return allEntities, allRelationships, nil
}

// scanDirectory recursively scans a directory for code files
func (cp *CodeProcessor) scanDirectory(dirPath string) ([]graph.CodeFile, error) {
	var files []graph.CodeFile

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			// Skip common directories that shouldn't be analyzed
			// But don't skip the root directory even if it's "."
			if path != dirPath && cp.shouldSkipDirectory(d.Name()) {
				log.Printf("⏭️ Skipping directory: %s", path)
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		log.Printf("🔍 Checking file: %s (ext: %s)", path, ext)
		if cp.supportedExtensions[ext] {
			log.Printf("✅ Processing supported file: %s", path)
			file, err := cp.createCodeFile(path)
			if err != nil {
				log.Printf("⚠️ Failed to read file %s: %v", path, err)
				return nil // Continue processing other files
			}
			if file != nil {
				files = append(files, *file)
			}
		} else {
			log.Printf("⏭️ Skipping unsupported file type: %s", path)
		}

		return nil
	})

	log.Printf("📊 Scanned directory, found %d supported files", len(files))
	return files, err
}

// shouldSkipDirectory determines if a directory should be skipped
func (cp *CodeProcessor) shouldSkipDirectory(dirName string) bool {
	// Don't skip current directory
	if dirName == "." {
		return false
	}

	skipDirs := map[string]bool{
		"node_modules": true,
		".git":         true,
		".svn":         true,
		"dist":         true,
		"build":        true,
		"out":          true,
		"target":       true,
		"bin":          true,
		"obj":          true,
		".vscode":      true,
		".idea":        true,
		"__pycache__":  true,
		"coverage":     true,
		".nyc_output":  true,
		"tmp":          true,
		"temp":         true,
		"logs":         true,
		"vendor":       true, // Go vendor directory
	}

	return skipDirs[dirName] || strings.HasPrefix(dirName, ".")
}

// createCodeFile creates a graph.CodeFile from a file path
func (cp *CodeProcessor) createCodeFile(filePath string) (*graph.CodeFile, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	language := cp.languageMap[ext]
	if language == "" {
		language = "unknown"
	}

	return &graph.CodeFile{
		Path:         filePath,
		Name:         filepath.Base(filePath),
		Extension:    ext,
		Content:      string(content),
		Language:     language,
		Size:         stat.Size(),
		LastModified: stat.ModTime(),
	}, nil
}

// extractDirectories extracts unique directories from file paths
func (cp *CodeProcessor) extractDirectories(files []graph.CodeFile) []string {
	directories := make(map[string]bool)

	for _, file := range files {
		dir := filepath.Dir(file.Path)
		for dir != "." && dir != "/" && dir != "" {
			directories[dir] = true
			dir = filepath.Dir(dir)
		}
	}

	var result []string
	for dir := range directories {
		result = append(result, dir)
	}

	return result
}

// createDirectoryEntity creates an entity for a directory
func (cp *CodeProcessor) createDirectoryEntity(dirPath, rootPath string) graph.Entity {
	relativePath := strings.TrimPrefix(dirPath, rootPath)
	relativePath = strings.TrimPrefix(relativePath, "/")

	baseName := filepath.Base(dirPath)
	if baseName == "." || baseName == "" {
		baseName = "root"
	}

	return graph.CreateEntity(baseName, graph.EntityTypeDirectory, graph.Properties{
		"path":         dirPath,
		"relativePath": relativePath,
		"fullPath":     dirPath,
	})
}

// analyzeFile analyzes a single code file
func (cp *CodeProcessor) analyzeFile(file graph.CodeFile) ([]graph.Entity, []graph.Relationship, error) {
	fileEntity := cp.createFileEntity(file)

	analyzer := cp.analyzerRegistry.GetAnalyzer(file.Language)
	return analyzer.Analyze(file, fileEntity)
}

// createFileEntity creates an entity for a file
func (cp *CodeProcessor) createFileEntity(file graph.CodeFile) graph.Entity {
	lineCount := len(strings.Split(file.Content, "\n"))

	return graph.CreateEntity(file.Name, graph.EntityTypeFile, graph.Properties{
		"path":         file.Path,
		"extension":    file.Extension,
		"language":     file.Language,
		"size":         file.Size,
		"lastModified": file.LastModified.Format(time.RFC3339),
		"lineCount":    lineCount,
	})
}

// createFileDirectoryRelationships creates relationships between files and directories
func (cp *CodeProcessor) createFileDirectoryRelationships(file graph.CodeFile, allEntities []graph.Entity) []graph.Relationship {
	var relationships []graph.Relationship
	fileDir := filepath.Dir(file.Path)

	// Find the directory entity
	var dirEntity *graph.Entity
	var fileEntity *graph.Entity

	for i := range allEntities {
		entity := &allEntities[i]
		if entity.Type == graph.EntityTypeDirectory {
			if path, ok := entity.Properties["path"].(string); ok && path == fileDir {
				dirEntity = entity
			}
		}
		if entity.Type == graph.EntityTypeFile {
			if path, ok := entity.Properties["path"].(string); ok && path == file.Path {
				fileEntity = entity
			}
		}
	}

	if dirEntity != nil && fileEntity != nil {
		relationships = append(relationships, graph.CreateRelationship(
			dirEntity.ID, fileEntity.ID, graph.RelationshipTypeContains, nil))
	}

	return relationships
}

// createImportRelationships creates relationships between imports and modules
func (cp *CodeProcessor) createImportRelationships(entities []graph.Entity) []graph.Relationship {
	var relationships []graph.Relationship

	// This is a simplified approach - in a real implementation,
	// you'd want to resolve imports to actual entities
	var importEntities []graph.Entity
	var moduleEntities []graph.Entity

	for _, entity := range entities {
		if entity.Type == graph.EntityTypeImport {
			importEntities = append(importEntities, entity)
		}
		if entity.Type == graph.EntityTypeFile || entity.Type == graph.EntityTypeModule {
			moduleEntities = append(moduleEntities, entity)
		}
	}

	for _, importEntity := range importEntities {
		if source, ok := importEntity.Properties["source"].(string); ok {
			// Try to find the corresponding module/file
			for _, moduleEntity := range moduleEntities {
				if strings.Contains(moduleEntity.Label, source) {
					if path, ok := moduleEntity.Properties["path"].(string); ok {
						if strings.Contains(path, source) {
							relationships = append(relationships, graph.CreateRelationship(
								importEntity.ID, moduleEntity.ID, graph.RelationshipTypeReferences, nil))
							break
						}
					}
				}
			}
		}
	}

	return relationships
}

// ProcessSingleFile processes a single code file and returns entities and relationships
func (cp *CodeProcessor) ProcessSingleFile(filePath string) ([]graph.Entity, []graph.Relationship, error) {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get file info for %s: %w", filePath, err)
	}

	// Determine language from extension
	ext := strings.ToLower(filepath.Ext(filePath))
	language := cp.languageMap[ext]
	if language == "" {
		language = "unknown"
	}

	// Create graph.CodeFile struct
	codeFile := graph.CodeFile{
		Path:         filePath,
		Name:         filepath.Base(filePath),
		Extension:    ext,
		Content:      string(content),
		Language:     language,
		Size:         fileInfo.Size(),
		LastModified: fileInfo.ModTime(),
	}

	// Analyze the file based on its language
	entities, relationships, err := cp.analyzeFile(codeFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to analyze file %s: %w", filePath, err)
	}

	// Create file entity and add it to the beginning
	fileEntity := graph.CreateEntity(codeFile.Name, graph.EntityTypeFile, graph.Properties{
		"path":         codeFile.Path,
		"extension":    codeFile.Extension,
		"language":     codeFile.Language,
		"size":         codeFile.Size,
		"lastModified": codeFile.LastModified,
	})

	// Combine file entity with analyzed entities
	allEntities := make([]graph.Entity, 0, len(entities)+1)
	allEntities = append(allEntities, fileEntity)
	allEntities = append(allEntities, entities...)

	return allEntities, relationships, nil
}
