[![OSS Lifecycle](https://img.shields.io/osslifecycle/honeycombio/hpsf?color/yellow)](https://github.com/honeycombio/home/blob/main/honeycomb-oss-lifecycle-and-practices.md)
[![GoDoc](https://godoc.org/github.com/honeycombio/hpsf?status.svg)](https://godoc.org/github.com/honeycombio/hpsf)

# HPSF -- EXPERIMENTAL!

## What it is

HPSF is an experimental format for a configuration language.

It will undergo radical changes for a while; please don't depend on it yet.

# hpsf

Here are some sample commands:

* go run ./cmd/hpsf -i examples/hpsf.yaml validate
* go run ./cmd/hpsf -i examples/hpsf.yaml rRules
* go run ./cmd/hpsf -i examples/hpsf.yaml rConfig

Here's an example that exercises a separate data table:

`go run ./cmd/hpsf -d API_Key=hello -i examples/hpsf2.yaml rConfig`