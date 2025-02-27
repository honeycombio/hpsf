package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/honeycombio/hpsf/pkg/data"
	"github.com/jessevdk/go-flags"
)

// hpsf2db is a command line tool that reads a set of components and templates
// from the embedded filesystem (the data package), and writes a set of
// components and templates to the the database. It is used in a deploy pipeline
// to ensure that the latest version of the components and templates are always
// available in the database.

type Options struct {
	VerifyCheckums     bool `short:"v" long:"verify-checksums" description:"verify checksums in stdin against the embedded templates and exit"`
	CalculateChecksums bool `short:"x" long:"calculate-checksum" description:"calculate checksums for the specified data, print to stdout, and exit"`
}

func main() {
	// Parse the command line arguments
	cmdopts := &Options{}
	parser := flags.NewParser(cmdopts, flags.Default)

	// read the command line and envvars into cmdargs
	cmdargs, err := parser.Parse()
	if err != nil {
		switch flagsErr := err.(type) {
		case *flags.Error:
			if flagsErr.Type == flags.ErrHelp {
				os.Exit(0)
			}
		}
		log.Fatalf("error reading command line: %v", err)
	}

	if len(cmdargs) == 0 {
		log.Fatalf("no command specified -- valid commands are 'templates' and 'components' (or both)")
	}

	for _, cmd := range cmdargs {
		checksums, err := data.CalculateChecksums(cmd)
		if err != nil {
			log.Fatalf("error calculating checksums from embedded data: %v", err)
		}
		if cmdopts.CalculateChecksums {
			// we print them checksum first because that's what the sha1sum command does
			for k, v := range checksums {
				fmt.Printf("%s  %s\n", v, k)
			}
		}
		if cmdopts.VerifyCheckums {
			// read checksums from stdin in the format "sha1  filename"
			// and verify that they match the embedded templates

			inputChecksums := make(map[string]string)
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				log.Fatalf("error reading checksums from stdin: %v", err)
			}
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if line == "" {
					continue
				}
				splits := strings.Split(line, "  ")
				if len(splits) != 2 {
					log.Fatalf("invalid checksum line: %s", line)
				}
				inputChecksums[splits[1]] = splits[0]
			}

			// we only compare the ones that we read from this directory (stdin might have extra files)
			for k, v := range checksums {
				ck, ok := inputChecksums[k]
				if !ok {
					log.Fatalf("checksum for %s not found in stdin -- new file?", k)
				}
				if ck != v {
					log.Fatalf("checksum mismatch for %s: %s != %s", k, ck, v)
				}
			}
		}
	}
}
