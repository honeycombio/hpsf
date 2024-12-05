package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/honeycombio/hpsf/pkg/config"
	"github.com/honeycombio/hpsf/pkg/hpsf"
	"github.com/honeycombio/hpsf/pkg/translator"
	"github.com/honeycombio/hpsf/pkg/validator"
	"github.com/honeycombio/hpsf/pkg/yaml"
	"github.com/jessevdk/go-flags"
	y "gopkg.in/yaml.v3"
)

type Options struct {
	Verbose bool     `short:"v" long:"verbose" description:"enable verbose mode"`
	Input   string   `short:"i" long:"input" description:"input file" default:"-"`
	Output  string   `short:"o" long:"output" description:"output file" default:"-"`
	Subs    []string `short:"s" long:"sub" description:"substitutions in the form 'context.varname=value'; can be repeated"`
}

func main() {
	// Parse the command line arguments
	cmdopts := &Options{}
	parser := flags.NewParser(cmdopts, flags.Default)

	// read the command line and envvars into cmdargs
	cmds, err := parser.Parse()
	if err != nil {
		switch flagsErr := err.(type) {
		case *flags.Error:
			if flagsErr.Type == flags.ErrHelp {
				os.Exit(0)
			}
		}
		log.Fatalf("error reading command line: %v", err)
	}

	if len(cmds) == 0 {
		log.Fatalf("no command specified")
	}

	// Read the input file
	inputData, err := readInput(cmdopts.Input)
	if err != nil {
		log.Fatalf("error reading input file: %v", err)
	}

	// Process the substitutions
	subst := translator.NewSubstitutor()
	for _, s := range cmdopts.Subs {
		splits := strings.Split(s, "=")
		if len(splits) != 2 {
			log.Fatalf("invalid substitution: %s", s)
		}
		kc, v := splits[0], splits[1]
		if !strings.Contains(kc, ".") {
			log.Fatalf("invalid substitution missing context: %s", s)
		}
		splits = strings.Split(kc, ".")
		c, k := splits[0], splits[1]

		subst.AddSubstitution(c, k, v)
	}
	subst.SetPriority("team", 3)
	subst.SetPriority("installation", 2)
	subst.SetPriority("cluster", 1)
	input := subst.DoSubstitutions(string(inputData))
	inputRdr := strings.NewReader(input)

	// Create the output file
	var outf io.Writer
	if cmdopts.Output == "-" {
		outf = os.Stdout
	} else {
		f, err := os.Create(cmdopts.Output)
		if err != nil {
			log.Fatalf("error creating output file: %v", err)
		}
		outf = f
		defer f.Close()
	}

	// create a translator
	tr := translator.NewTranslator()

	switch cmds[0] {
	case "format":
		hpsf, err := unmarshalHPSF(inputRdr)
		if err != nil {
			log.Fatalf("error unmarshaling input file: %v", err)
		}
		// write it to the output file as yaml
		data, err := y.Marshal(hpsf)
		if err != nil {
			log.Fatalf("error marshaling output file: %v", err)
		}
		_, err = outf.Write(data)
		if err != nil {
			log.Fatalf("error writing output file: %v", err)
		}
	case "validate":
		// validate the input file
		_, err := validator.EnsureYAML(inputData)
		if err != nil {
			log.Fatalf("error validating input file: %v", err)
		}

		var hpsf hpsf.HPSF
		err = y.Unmarshal(inputData, &hpsf)
		if err != nil {
			log.Fatalf("error unmarshaling to HPSF: %v", err)
		}

		// validate the HPSF
		errors := hpsf.Validate()
		if len(errors) > 0 {
			for _, e := range errors {
				log.Printf("error: %v", e)
			}
			os.Exit(1)
		}

		log.Printf("HPSF is valid")

	case "dotify":
		// create a dotted config from the input file and write it to the output
		m := make(map[string]any)
		rdr := bytes.NewReader(inputData)
		dec := y.NewDecoder(rdr)
		err := dec.Decode(&m)
		if err != nil {
			log.Fatalf("error unmarshaling to yaml: %v", err)
		}
		dc := yaml.NewDottedConfig(m)
		for k, v := range dc {
			fmt.Fprintf(outf, "%s: %v\n", k, v)
		}
		os.Exit(0)

	case "rConfig", "rRules", "cConfig":
		hpsf, err := unmarshalHPSF(inputRdr)
		if err != nil {
			log.Fatalf("error unmarshaling input file: %v", err)
		}
		var ct config.Type
		switch cmds[0] {
		case "rConfig":
			ct = config.RefineryConfigType
		case "rRules":
			ct = config.RefineryRulesType
		case "cConfig":
			ct = config.CollectorConfigType
		}
		cfg, err := tr.GenerateConfig(hpsf, ct)
		if err != nil {
			log.Fatalf("error translating refinery config: %v", err)
		}
		data, _, err := cfg.RenderYAML()
		if err != nil {
			log.Fatalf("error marshaling output file: %v", err)
		}
		_, err = outf.Write(data)
		if err != nil {
			log.Fatalf("error writing output file: %v", err)
		}
	default:
		log.Fatalf("unknown command: %s", cmds[0])
	}
}

func readInput(filename string) ([]byte, error) {
	// Open the fIn
	var fIn io.Reader
	if (filename == "") || (filename == "-") {
		fIn = os.Stdin
	} else {
		f, err := os.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("error opening file %s: %v", filename, err)
		}
		fIn = f
		defer f.Close()
	}

	// read it into a buffer
	data, err := io.ReadAll(fIn)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %v", filename, err)
	}
	return data, nil
}

func unmarshalHPSF(data io.Reader) (*hpsf.HPSF, error) {
	var hpsf hpsf.HPSF
	dec := y.NewDecoder(data)
	err := dec.Decode(&hpsf)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling to yaml: %v", err)
	}
	return &hpsf, nil
}
