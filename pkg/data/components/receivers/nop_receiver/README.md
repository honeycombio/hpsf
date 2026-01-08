# Receive Nothing

**Kind:** `NopReceiver` | **Version:** `v0.1.0` | **Status:** development

## Overview

description: A simple no-op receiver. This receiver does nothing. It is required for the minimal configuration.tags:- category:input- service:collector- signal:OTelTraces

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-nop_receiver
    kind: NopReceiver
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
