package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jessevdk/go-flags"
	"gopkg.in/yaml.v3"
)

type Options struct {
	Export string   `short:"e" long:"export" description:"Export components to CSV file (use - for stdout)"`
	Import string   `short:"i" long:"import" description:"Import CSV file and apply changes to components (use - for stdin)"`
	Styles []string `short:"s" long:"style" description:"Component style to include (can be repeated)"`
	Fields []string `short:"f" long:"field" description:"Top-level field to include (can be repeated)"`
}

type Component struct {
	node     *yaml.Node
	filePath string
}

func main() {
	opts := &Options{}
	parser := flags.NewParser(opts, flags.Default)
	parser.ShortDescription = "Component property CSV export/import tool"
	parser.LongDescription = `This tool allows you to extract key fields from component YAML files to a CSV file for editing in a spreadsheet, then apply changes back to the YAML files.

The tool performs two main operations:
• Export: Extract component fields to CSV (one row per component)
• Import: Read CSV and modify components based on values

Multi-line fields are unwrapped (newlines replaced with spaces) when exported to CSV.
On import, text longer than 80 characters is word-wrapped before being written to YAML.

The 'kind' field is always included as the first column for matching components.
If no fields are specified, defaults to: name, style, status, version, summary, description.
If no files are specified, processes all *.yaml files in the current directory.

Examples:
  propdump --export components.csv --style condition
  propdump --import components.csv
  propdump --export - --field name --field status | head -10`

	args, err := parser.Parse()
	if err != nil {
		switch flagsErr := err.(type) {
		case *flags.Error:
			if flagsErr.Type == flags.ErrHelp {
				os.Exit(0)
			}
		}
		log.Fatalf("error reading command line: %v", err)
	}

	if opts.Import != "" {
		processInput(opts, args)
	} else if opts.Export != "" {
		processOutput(opts, args)
	} else {
		log.Fatal("Must specify either --import or --export")
	}
}

func processOutput(opts *Options, args []string) {
	// Set default fields if none specified
	fields := opts.Fields
	if len(fields) == 0 {
		fields = []string{"name", "style", "status", "version", "tags", "summary", "description"}
	}

	// Set default files if none specified
	files := args
	if len(files) == 0 {
		matches, err := filepath.Glob("*.yaml")
		if err != nil {
			log.Fatal(err)
		}
		files = matches
	}

	components := loadComponents(files)
	filteredComponents := filterComponents(components, opts.Styles)

	var output io.Writer
	if opts.Export == "-" {
		output = os.Stdout
	} else {
		file, err := os.Create(opts.Export)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		output = file
	}

	writer := csv.NewWriter(output)
	defer writer.Flush()

	// Write header
	header := append([]string{"kind"}, fields...)
	writer.Write(header)

	// Write data
	for _, comp := range filteredComponents {
		if kind := getFieldFromNode(comp.node, "kind"); kind != "" {
			row := []string{kind}
			for _, field := range fields {
				value := getFieldFromNode(comp.node, field)
				// Unwrap multiline text: replace all whitespace around line endings with single space
				value = unwrapMultilineText(value)
				row = append(row, value)
			}
			writer.Write(row)
		}
	}

	if opts.Export != "-" {
		fmt.Fprintf(os.Stderr, "Exported %d components to %s\n", len(filteredComponents), opts.Export)
	}
}

func processInput(opts *Options, args []string) {
	// Set default files if none specified
	files := args
	if len(files) == 0 {
		matches, err := filepath.Glob("*.yaml")
		if err != nil {
			log.Fatal(err)
		}
		files = matches
	}

	components := loadComponents(files)
	componentMap := make(map[string]*Component)
	for i := range components {
		if kind := getFieldFromNode(components[i].node, "kind"); kind != "" {
			componentMap[kind] = &components[i]
		}
	}

	var input io.Reader
	if opts.Import == "-" {
		input = os.Stdin
	} else {
		file, err := os.Open(opts.Import)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		input = file
	}

	reader := csv.NewReader(input)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	if len(records) == 0 {
		log.Fatal("CSV file is empty")
	}

	header := records[0]
	if len(header) == 0 || header[0] != "kind" {
		log.Fatal("CSV file must have 'kind' as first column")
	}

	modifiedCount := 0
	errorCount := 0

	for i, record := range records[1:] {
		if len(record) != len(header) {
			log.Printf("Row %d: incorrect number of columns", i+2)
			errorCount++
			continue
		}

		kind := record[0]
		comp, exists := componentMap[kind]
		if !exists {
			log.Printf("Row %d: component kind '%s' not found", i+2, kind)
			errorCount++
			continue
		}

		modified := false
		for j, field := range header[1:] {
			if j+1 >= len(record) {
				break
			}
			newValue := record[j+1]

			// Wrap long lines to 100 characters
			if len(newValue) > 100 {
				newValue = wrapText(newValue, 100)
			}

			if setFieldInNode(comp.node, field, newValue) {
				modified = true
			}
		}

		if modified {
			if err := saveComponent(comp); err != nil {
				log.Printf("Error saving component %s: %v", getFieldFromNode(comp.node, "kind"), err)
				errorCount++
			} else {
				modifiedCount++
			}
		}
	}

	fmt.Fprintf(os.Stderr, "Modified %d components\n", modifiedCount)

	if errorCount > 0 {
		os.Exit(1)
	}
}

func loadComponents(filePaths []string) []Component {
	var components []Component

	for _, filePath := range filePaths {
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Error reading %s: %v", filePath, err)
			continue
		}

		var node yaml.Node
		if err := yaml.Unmarshal(data, &node); err != nil {
			log.Printf("Error parsing %s: %v", filePath, err)
			continue
		}

		comp := Component{
			node:     &node,
			filePath: filePath,
		}
		components = append(components, comp)
	}

	return components
}

func filterComponents(components []Component, styles []string) []Component {
	if len(styles) == 0 {
		return components
	}

	styleSet := make(map[string]bool)
	for _, style := range styles {
		styleSet[style] = true
	}

	var filtered []Component
	for _, comp := range components {
		if style := getFieldFromNode(comp.node, "style"); styleSet[style] {
			filtered = append(filtered, comp)
		}
	}

	return filtered
}

func getFieldFromNode(node *yaml.Node, field string) string {
	if node.Kind != yaml.DocumentNode || len(node.Content) == 0 {
		return ""
	}

	mappingNode := node.Content[0]
	if mappingNode.Kind != yaml.MappingNode {
		return ""
	}

	for i := 0; i < len(mappingNode.Content); i += 2 {
		keyNode := mappingNode.Content[i]
		valueNode := mappingNode.Content[i+1]

		if keyNode.Kind == yaml.ScalarNode && keyNode.Value == field {
			if valueNode.Kind == yaml.ScalarNode {
				return valueNode.Value
			} else if valueNode.Kind == yaml.SequenceNode {
				// Handle arrays - serialize as []item1,item2,item3
				var items []string
				for _, itemNode := range valueNode.Content {
					if itemNode.Kind == yaml.ScalarNode {
						items = append(items, itemNode.Value)
					}
				}
				return "[]" + strings.Join(items, ",")
			}
		}
	}

	return ""
}

func setFieldInNode(node *yaml.Node, field, value string) bool {
	if node.Kind != yaml.DocumentNode || len(node.Content) == 0 {
		return false
	}

	mappingNode := node.Content[0]
	if mappingNode.Kind != yaml.MappingNode {
		return false
	}

	currentValue := getFieldFromNode(node, field)
	if currentValue == value {
		return false
	}

	for i := 0; i < len(mappingNode.Content); i += 2 {
		keyNode := mappingNode.Content[i]
		valueNode := mappingNode.Content[i+1]

		if keyNode.Kind == yaml.ScalarNode && keyNode.Value == field {
			if valueNode.Kind == yaml.ScalarNode {
				// Check if this is an array format
				if strings.HasPrefix(value, "[]") {
					// Convert to sequence node
					valueNode.Kind = yaml.SequenceNode
					valueNode.Style = 0
					valueNode.Value = ""

					// Parse array items
					arrayContent := value[2:] // Remove "[]" prefix
					if arrayContent != "" {
						items := strings.Split(arrayContent, ",")
						valueNode.Content = make([]*yaml.Node, len(items))
						for j, item := range items {
							valueNode.Content[j] = &yaml.Node{
								Kind:  yaml.ScalarNode,
								Value: strings.TrimSpace(item),
							}
						}
					} else {
						valueNode.Content = []*yaml.Node{}
					}
				} else {
					valueNode.Value = value
					// Handle multiline strings
					if strings.Contains(value, "\n") {
						valueNode.Style = yaml.LiteralStyle
					} else {
						valueNode.Style = 0
					}
				}
				return true
			} else if valueNode.Kind == yaml.SequenceNode {
				// Check if this is an array format
				if strings.HasPrefix(value, "[]") {
					// Parse array items
					arrayContent := value[2:] // Remove "[]" prefix
					if arrayContent != "" {
						items := strings.Split(arrayContent, ",")
						valueNode.Content = make([]*yaml.Node, len(items))
						for j, item := range items {
							valueNode.Content[j] = &yaml.Node{
								Kind:  yaml.ScalarNode,
								Value: strings.TrimSpace(item),
							}
						}
					} else {
						valueNode.Content = []*yaml.Node{}
					}
					return true
				}
			}
		}
	}

	return false
}

func unwrapMultilineText(text string) string {
	if text == "" {
		return text
	}

	// Replace all whitespace around line endings with a single space
	// This handles cases like:
	//   "word1\n  word2" -> "word1 word2"
	//   "word1  \n  word2" -> "word1 word2"
	//   "word1\n\nword2" -> "word1 word2"
	re := regexp.MustCompile(`\s*\n\s*`)
	unwrapped := re.ReplaceAllString(text, " ")

	// Clean up any multiple spaces and trim
	unwrapped = regexp.MustCompile(`\s+`).ReplaceAllString(unwrapped, " ")
	return strings.TrimSpace(unwrapped)
}

func wrapText(text string, width int) string {
	if len(text) <= width {
		return text
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var result strings.Builder
	lineLength := 0

	for i, word := range words {
		if i > 0 {
			if lineLength+1+len(word) > width {
				result.WriteString("\n")
				lineLength = 0
			} else {
				result.WriteString(" ")
				lineLength++
			}
		}
		result.WriteString(word)
		lineLength += len(word)
	}

	return result.String()
}

func saveComponent(comp *Component) error {
	// Use encoder to preserve original formatting better
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2) // Use 2-space indentation to match original

	if err := encoder.Encode(comp.node); err != nil {
		return err
	}

	if err := encoder.Close(); err != nil {
		return err
	}

	return os.WriteFile(comp.filePath, buf.Bytes(), 0644)
}
