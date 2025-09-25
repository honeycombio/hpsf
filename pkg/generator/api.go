package generator

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/honeycombio/hpsf/pkg/hpsf"
)

// GenerateFromFile generates an HPSF workflow from a Refinery rules file
func GenerateFromFile(rulesPath string) (*hpsf.HPSF, error) {
	// Read the rules file
	rulesData, err := os.ReadFile(rulesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read rules file %s: %w", rulesPath, err)
	}

	// Generate the workflow
	generator := NewGenerator()
	return generator.GenerateWorkflow(rulesData)
}

// GenerateFromDirectory generates an HPSF workflow from a directory containing Refinery rules
// It looks for files with common Refinery naming patterns
func GenerateFromDirectory(dirPath string) (*hpsf.HPSF, error) {
	rulesPath, err := findRefineryRulesFile(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to find Refinery rules file in directory %s: %w", dirPath, err)
	}

	return GenerateFromFile(rulesPath)
}

// GenerateFromBytes generates an HPSF workflow from raw rules data
func GenerateFromBytes(rulesData []byte) (*hpsf.HPSF, error) {
	generator := NewGenerator()
	return generator.GenerateWorkflow(rulesData)
}

// WriteWorkflowToFile writes an HPSF workflow to a file in YAML format
func WriteWorkflowToFile(workflow *hpsf.HPSF, outputPath string) error {
	yamlContent, err := workflow.AsYAML()
	if err != nil {
		return fmt.Errorf("failed to convert workflow to YAML: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(yamlContent), 0644); err != nil {
		return fmt.Errorf("failed to write workflow to file %s: %w", outputPath, err)
	}

	return nil
}

// findRefineryRulesFile looks for Refinery rules files in a directory
func findRefineryRulesFile(dirPath string) (rulesPath string, err error) {
	var possibleRulesFiles []string

	err = filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		name := d.Name()
		lowerName := filepath.Base(name)

		// Look for rules files
		if isRulesFile(lowerName) {
			possibleRulesFiles = append(possibleRulesFiles, path)
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	// Select the most likely rules file
	if len(possibleRulesFiles) == 0 {
		return "", fmt.Errorf("no Refinery rules files found in directory")
	}
	rulesPath = selectBestMatch(possibleRulesFiles, []string{"rules.yaml", "refinery-rules.yaml", "sampling.yaml", "rules.yml"})

	return rulesPath, nil
}


// isRulesFile checks if a filename looks like a Refinery rules file
func isRulesFile(filename string) bool {
	rulesPatterns := []string{
		"rules.yaml", "rules.yml",
		"refinery-rules.yaml", "refinery-rules.yml",
		"sampling.yaml", "sampling.yml",
		"sampling-rules.yaml", "sampling-rules.yml",
	}

	for _, pattern := range rulesPatterns {
		if filename == pattern {
			return true
		}
	}

	// Also check if it contains "rules" or "sampling" and is a YAML file
	return (containsWord(filename, "rules") || containsWord(filename, "sampling")) && isYAMLFile(filename)
}

// containsWord checks if a string contains a specific word
func containsWord(s, word string) bool {
	return filepath.Base(s) == word ||
		   filepath.Base(s) == word+".yaml" ||
		   filepath.Base(s) == word+".yml" ||
		   len(s) >= len(word) && s[:len(word)] == word ||
		   len(s) >= len(word) && s[len(s)-len(word):] == word
}

// isYAMLFile checks if a filename has a YAML extension
func isYAMLFile(filename string) bool {
	ext := filepath.Ext(filename)
	return ext == ".yaml" || ext == ".yml"
}

// selectBestMatch selects the best matching file from a list based on preferred patterns
func selectBestMatch(files []string, preferredPatterns []string) string {
	// First, try to find exact matches with preferred patterns
	for _, pattern := range preferredPatterns {
		for _, file := range files {
			if filepath.Base(file) == pattern {
				return file
			}
		}
	}

	// If no exact match, return the first file
	return files[0]
}

// ValidateInputFile validates that the provided rules file exists and is readable
func ValidateInputFile(rulesPath string) error {
	if _, err := os.Stat(rulesPath); err != nil {
		return fmt.Errorf("rules file %s is not accessible: %w", rulesPath, err)
	}

	return nil
}