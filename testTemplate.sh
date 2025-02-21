#! /usr/bin/env bash

# Generates configs for a given template, placing in tmp/ directory
	# $1 - template

# check if $1.yaml exists
if [ ! -f $1.yaml ]; then
    echo "File $1.yaml does not exist"
    exit 1
fi

mkdir -p tmp/

# Get just the filename without the path
BASENAME=$(basename $1)

go run ./cmd/hpsf -i $1.yaml -o tmp/${BASENAME}.cconfig.yaml cConfig
go run ./cmd/hpsf -i $1.yaml -o tmp/${BASENAME}.rconfig.yaml rConfig
go run ./cmd/hpsf -i $1.yaml -o tmp/${BASENAME}.rrules.yaml rRules
