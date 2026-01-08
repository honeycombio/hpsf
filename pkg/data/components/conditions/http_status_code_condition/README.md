# Check HTTP Status Code

**Kind:** `HTTPStatusCodeCondition` | **Version:** `v0.1.0` | **Status:** development

## Overview

description: |-Sample based on the http status of processes within a trace. If any spans of a trace have a status code within the selected min and max, the condition will evaluate to true. tags:- category:refinery_rule- service:refinery

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-h_t_t_p_status_code_condition
    kind: HTTPStatusCodeCondition
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
