# Check HTTP Status

**Kind:** `HTTPStatusCondition` | **Version:** `v0.1.0` | **Status:** alpha

## Overview

description: |-Samples HTTP errors based on the http status of processes within the trace. If any span a trace hasa status code in the 500s, it will be sampled at the error rate. If any span has a status code inthe 400s, it will be sampled at the user error rate. All other traces will be sampled at the defaultrate.

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-h_t_t_p_status_condition
    kind: HTTPStatusCondition
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
