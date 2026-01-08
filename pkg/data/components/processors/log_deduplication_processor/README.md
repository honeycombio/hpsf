# Deduplicate Logs

**Kind:** `LogDeduplicationProcessor` | **Version:** `v0.1.0` | **Status:** development

## Overview

description: |-This processor detects identical logs over a range of time (60 seconds by default) and emitsa single log with the count of logs that were deduplicated (reported as sampleRate by default).tags:- category:processor

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-log_deduplication_processor
    kind: LogDeduplicationProcessor
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
