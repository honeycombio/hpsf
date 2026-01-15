# Symbolicate Swift/iOS Errors

**Kind:** `DsymSymbolicationProcessor` | **Version:** `v0.1.0` | **Status:** development

## Overview

description: |Symbolicate iOS/Swift stack traces using dSYM symbol files.Supports dSYM files stored in AWS S3 or Google Cloud Storage.tags:- category:processor

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-dsym_symbolication_processor
    kind: DsymSymbolicationProcessor
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
