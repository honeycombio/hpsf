# Symbolicate Java/Android Errors

**Kind:** `ProguardSymbolicationProcessor` | **Version:** `v0.1.0` | **Status:** development

## Overview

description: |Symbolicate Java/Kotlin stack traces from Android applications using ProGuard mapping files.Supports ProGuard mapping files stored in AWS S3 or Google Cloud Storage.tags:- category:processor

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-proguard_symbolication_processor
    kind: ProguardSymbolicationProcessor
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
