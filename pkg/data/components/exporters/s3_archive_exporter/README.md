# Send to S3 Archive

**Kind:** `S3ArchiveExporter` | **Version:** `v0.1.0` | **Status:** alpha

## Overview

description: Sends the telemetry S3 for long-term storage to the location you choose.tags:- category:output- service:collector- signal:OTelTraces

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-s3_archive_exporter
    kind: S3ArchiveExporter
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
