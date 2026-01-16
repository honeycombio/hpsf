# Migration Guide: HTTPStatusCondition

## HTTPStatusCondition → HTTPStatusCodeCondition

**Status:** HTTPStatusCondition deprecated 2026-01-15
**Released:** 2026-01-15
**Replacement:** [HTTPStatusCodeCondition](../http_status_code_condition/README.md)

### Breaking Changes

1. **Component Kind Change:** `HTTPStatusCondition` → `HTTPStatusCodeCondition`
   - **Impact:** All configs using HTTPStatusCondition must update component kind
   - **Reason:** More explicit naming and improved range-based filtering

2. **Property Changes:** `Operator` + `Value` → `Minimum` + `Maximum`
   - **Impact:** Must specify range instead of single operator/value
   - **Reason:** HTTPStatusCodeCondition uses inclusive ranges for clearer status code filtering

3. **New Property:** `Fields` (default: `["http.status_code", "http.response.status_code"]`)
   - **Impact:** Optional - defaults match HTTPStatusCondition behavior
   - **Reason:** Allow customization of which fields to check

### Migration Steps

**Manual:**

Before (HTTPStatusCondition with >= 400):
```yaml
components:
  - name: check-errors
    kind: HTTPStatusCondition
    properties:
      - name: Operator
        value: '>='
      - name: Value
        value: 400
```

After (HTTPStatusCodeCondition):
```yaml
components:
  - name: check-errors
    kind: HTTPStatusCodeCondition
    properties:
      - name: Minimum
        value: 400
      - name: Maximum
        value: 599
```

### Operator Migration

| HTTPStatusCondition | HTTPStatusCodeCondition | Notes |
|--------------------|------------------------|-------|
| Operator: `>=`, Value: `400` | Minimum: `400`, Maximum: `599` | Check for 4xx and 5xx errors |
| Operator: `>=`, Value: `500` | Minimum: `500`, Maximum: `599` | Check for 5xx errors only |
| Operator: `=`, Value: `404` | Minimum: `404`, Maximum: `404` | Check for specific status |
| Operator: `<`, Value: `400` | Minimum: `100`, Maximum: `399` | Check for success codes |

### Examples

**Check for all errors (4xx and 5xx):**

Before:
```yaml
properties:
  - name: Operator
    value: '>='
  - name: Value
    value: 400
```

After:
```yaml
properties:
  - name: Minimum
    value: 400
  - name: Maximum
    value: 599
```

**Check for server errors only (5xx):**

Before:
```yaml
properties:
  - name: Operator
    value: '>='
  - name: Value
    value: 500
```

After:
```yaml
properties:
  - name: Minimum
    value: 500
  - name: Maximum
    value: 599
```

### Backward Compatibility

None. This is a complete replacement requiring config updates.

### Additional Capabilities

HTTPStatusCodeCondition adds:
- Range-based filtering with inclusive min/max bounds
- Customizable field names via `Fields` property
- More intuitive configuration for status code ranges
