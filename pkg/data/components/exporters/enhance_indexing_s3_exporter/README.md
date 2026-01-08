# Send to Enhance S3 Archive

**Kind:** `EnhanceIndexingS3Exporter` | **Version:** `v0.1.0` | **Status:** alpha

## Overview

description: |This component writes telemetry data to S3 while simultaneously generating field-based indexes that enable efficient querying of unsampled data. Automatically indexes trace.trace_id, service.name, and session.id, with support for additional custom fields.

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-enhance_indexing_s3_exporter
    kind: EnhanceIndexingS3Exporter
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
