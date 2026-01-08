# Mask sensitive values

**Kind:** `RedactionProcessor` | **Version:** `v0.1.0` | **Status:** development

## Overview

description: |-This processor is used to redact (mask) sensitive values in trace and logattributes based on predefined patterns or custom regular expressions. Ithelps ensure that sensitive information such as phone numbers, credit cardnumbers, and other personally identifiable information is not sent with telemetry.

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-redaction_processor
    kind: RedactionProcessor
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
