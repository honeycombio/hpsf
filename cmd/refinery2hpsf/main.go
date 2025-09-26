package main

import (
	"log"
	"os"

	"github.com/honeycombio/hpsf/pkg/generator"
	"github.com/honeycombio/hpsf/pkg/validator"
	"github.com/jessevdk/go-flags"
)

type Options struct {
	Verbose       bool   `short:"v" long:"verbose" description:"enable verbose mode"`
	Output        string `short:"o" long:"output" description:"output file" default:"-"`
	RefineryRules string `short:"r" long:"refinery-rules" description:"Path to Refinery rules file" required:"true"`
}

func main() {
	// Parse the command line arguments
	opts := &Options{}
	parser := flags.NewParser(opts, flags.Default)
	parser.Usage = "[OPTIONS]\n\nGenerate HPSF workflow from Refinery sampling rules"

	_, err := parser.Parse()
	if err != nil {
		switch flagsErr := err.(type) {
		case *flags.Error:
			if flagsErr.Type == flags.ErrHelp {
				os.Exit(0)
			}
		}
		log.Fatalf("error reading command line: %v", err)
	}

	// Validate input file exists
	if _, err := os.Stat(opts.RefineryRules); err != nil {
		log.Fatalf("rules file %s is not accessible: %w", opts.RefineryRules, err)
	}

	// Read the Refinery rules
	rulesData, err := os.ReadFile(opts.RefineryRules)
	if err != nil {
		log.Fatalf("failed to read rules file %s: %w", opts.RefineryRules, err)
	}

	// Generate the workflow
	gen := generator.NewGenerator()
	workflow, err := gen.GenerateWorkflow(rulesData)
	if err != nil {
		log.Fatalf("error generating HPSF workflow: %v", err)
	}

	// Validate the generated workflow
	if verrors := workflow.Validate(); verrors != nil {
		if hErr, ok := verrors.(validator.Result); ok {
			log.Printf("warning: generated workflow has validation errors: %v", hErr.Msg)
			for _, e := range hErr.Details {
				log.Printf("  warning: %v", e)
			}
		} else {
			log.Printf("warning: unexpected validation error: %v", verrors)
		}
	}

	// Determine output destination
	var output *os.File
	if opts.Output == "-" {
		output = os.Stdout
	} else {
		output, err = os.Create(opts.Output)
		if err != nil {
			log.Fatalf("error creating output file: %v", err)
		}
		defer output.Close()
	}

	// Write the workflow to output
	yamlContent, err := workflow.AsYAML()
	if err != nil {
		log.Fatalf("error converting workflow to YAML: %v", err)
	}

	_, err = output.Write([]byte(yamlContent))
	if err != nil {
		log.Fatalf("error writing output: %v", err)
	}

	if opts.Verbose {
		log.Printf("Successfully generated HPSF workflow '%s' with %d components and %d connections",
			workflow.Name, len(workflow.Components), len(workflow.Connections))
	}
}
