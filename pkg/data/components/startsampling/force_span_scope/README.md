# Set Span Scope

**Kind:** `ForceSpanScope` | **Version:** `v0.1.0` | **Status:** alpha

## Overview

description: |-Normally, the scope of a Refinery rule is "trace", which means that when the rule has multipleconditions, it will match when any condition is true for any span in the trace, even if theconditions are not related to the same span. This component forces the scope of the rule to be"span" rather than "trace", so that the rule will only match when all conditions are true for the

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-force_span_scope
    kind: ForceSpanScope
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
