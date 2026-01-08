# Symbolicate JavaScript/Web Errors

**Kind:** `SourcemapSymbolicationProcessor` | **Version:** `v0.1.0` | **Status:** development

## Overview

description: |Symbolicate JavaScript/TypeScript stack traces from web applications using source map files.Supports source maps stored in AWS S3 or Google Cloud Storage.tags:- category:processor

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-sourcemap_symbolication_processor
    kind: SourcemapSymbolicationProcessor
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
