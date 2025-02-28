package main

import (
	"fmt"
	"io"
	"log"
	"maps"
	"os"
	"slices"
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

func diffMaps(orig, update map[string]string) (added, removed, changed map[string]string) {
	added = make(map[string]string)
	removed = make(map[string]string)
	changed = make(map[string]string)

	for k, v := range orig {
		if _, ok := update[k]; !ok {
			removed[k] = v
		} else if update[k] != v {
			changed[k] = v
		}
	}

	for k, v := range update {
		if _, ok := orig[k]; !ok {
			added[k] = v
		}
	}

	return added, removed, changed
}

func main() {
	// Parse the command line arguments
	cmdopts := &Options{}
	parser := flags.NewParser(cmdopts, flags.Default)

	// read the command line
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

	checksums, err := data.CalculateChecksums()
	if err != nil {
		log.Fatalf("error calculating checksums from embedded data: %v", err)
	}

	if cmdopts.CalculateChecksums {
		// we print them checksum first because that's what the sha1sum command does
		// and we print them in sorted order so they're easily comparable
		keys := slices.Sorted(maps.Keys(checksums))
		for _, k := range keys {
			fmt.Printf("%s  %s\n", checksums[k], k)
		}
		// and we're done
		os.Exit(0)
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

		added, removed, changed := diffMaps(checksums, inputChecksums)
		if len(added) > 0 {
			fmt.Println("added:")
			for k, v := range added {
				fmt.Printf("%s  %s\n", v, k)
			}
		}
		if len(removed) > 0 {
			fmt.Println("removed:")
			for k, v := range removed {
				fmt.Printf("%s  %s\n", v, k)
			}
		}
		if len(changed) > 0 {
			fmt.Println("changed:")
			for k, v := range changed {
				fmt.Printf("%s  %s\n", v, k)
			}
		}
		if len(added) == 0 && len(removed) == 0 && len(changed) == 0 {
			fmt.Println("no changes")
			// we return 0 to indicate that the checksums match
			os.Exit(0)
		} else {
			// we return 1 to indicate that the checksums do not match
			os.Exit(1)
		}
	}
}
