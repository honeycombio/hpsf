package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/honeycombio/hpsf/pkg/config/tmpl"
	"github.com/honeycombio/hpsf/pkg/data"
	"github.com/honeycombio/hpsf/pkg/hpsf"
	"github.com/honeycombio/hpsf/pkg/hpsftypes"
	"github.com/honeycombio/hpsf/pkg/translator"
	"github.com/honeycombio/hpsf/pkg/validator"
	"github.com/jessevdk/go-flags"
	y "gopkg.in/yaml.v3"
)

type Options struct {
	Verbose bool     `short:"v" long:"verbose" description:"enable verbose mode"`
	Input   string   `short:"i" long:"input" description:"input file" default:"-"`
	Output  string   `short:"o" long:"output" description:"output file" default:"-"`
	Subs    []string `short:"s" long:"sub" description:"substitutions in the form 'context.varname=value'; can be repeated"`
	Data    []string `short:"d" long:"data" description:"data in the form 'key=value'; can be repeated"`
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

	// Process the data
	userdata := make(map[string]any)
	for _, d := range cmdopts.Data {
		splits := strings.Split(d, "=")
		if len(splits) != 2 {
			log.Fatalf("invalid data: %s", d)
		}
		userdata[splits[0]] = splits[1]
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

	// Create the output file
	var outf io.Writer
	var f *os.File
	if cmdopts.Output == "-" {
		outf = os.Stdout
	} else {
		f, err = os.Create(cmdopts.Output)
		if err != nil {
			log.Fatalf("error creating output file: %v", err)
		}
		outf = f
		defer f.Close()
	}

	// create a translator that knows about components
	tr := translator.NewEmptyTranslator()
	// for this command line app, we load the embedded components, but
	// a real app should load them from a database
	components, err := data.LoadEmbeddedComponents()
	if err != nil {
		if f != nil {
			f.Close()
		}
		log.Fatalf("error loading embedded components: %v", err)
	}
	// install the components
	tr.InstallComponents(components)

	switch cmds[0] {
	case "format":
		h, err := hpsf.FromYAML(input)
		if err != nil {
			log.Fatalf("error unmarshaling input file: %v", err)
		}
		// write it to the output file as yaml
		data, err := y.Marshal(&h)
		if err != nil {
			log.Fatalf("error marshaling output file: %v", err)
		}
		_, err = outf.Write(data)
		if err != nil {
			log.Fatalf("error writing output file: %v", err)
		}
	case "validate":
		err = hpsf.EnsureHPSFYAML(string(inputData))
		if err != nil {
			log.Fatalf("input file is not hpsf: %v", err)
		}

		var h hpsf.HPSF
		err = y.Unmarshal(inputData, &h)
		if err != nil {
			log.Fatalf("error unmarshaling to HPSF: %v", err)
		}

		// validate the HPSF
		if verrors := h.Validate(); verrors != nil {
			if hErr, ok := verrors.(validator.Result); ok {
				log.Printf("error: %v", hErr.Msg)
				for _, e := range hErr.Details {
					log.Printf("  error: %v", e)
				}
				os.Exit(1)
			}

			log.Printf("unexpected validation error: %v", verrors)
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
		dc := tmpl.NewDottedConfig(m)
		for k, v := range dc {
			fmt.Fprintf(outf, "%s: %v\n", k, v)
		}
		os.Exit(0)

	case "rConfig", "rRules", "cConfig":
		hpsf, err := hpsf.FromYAML(input)
		if err != nil {
			log.Fatalf("error unmarshaling input file: %v", err)
		}
		var ct hpsftypes.Type
		switch cmds[0] {
		case "rConfig":
			ct = hpsftypes.RefineryConfig
		case "rRules":
			ct = hpsftypes.RefineryRules
		case "cConfig":
			ct = hpsftypes.CollectorConfig
		}
		cfg, err := tr.GenerateConfig(&hpsf, ct, "latest", userdata)
		if err != nil {
			log.Fatalf("error translating config: %v", err)
		}
		data, err := cfg.RenderYAML()
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
			return nil, fmt.Errorf("error opening file %s: %w", filename, err)
		}
		fIn = f
		defer f.Close()
	}

	// read it into a buffer
	data, err := io.ReadAll(fIn)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filename, err)
	}
	return data, nil
}
