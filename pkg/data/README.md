# data package for HPSF

This directory contains the HPSF source code for the standard Honeycomb components and templates.

The live application fetches this information from a database, but we needed a place to store this data
in source form. The components here are available to applications as TemplateComponent objects by calling `LoadEmbeddedComponents()`.

The component file systems containing these directories are also exported in `data.go`.

