# Migration Guide: CompareStringFieldCondition

## CompareStringFieldCondition → StringValueCondition

**Status:** CompareStringFieldCondition deprecated 2026-01-15
**Released:** 2026-01-15
**Replacement:** [StringValueCondition](../string_value_condition/README.md)

### Breaking Changes

1. **Component Kind Change:** `CompareStringFieldCondition` → `StringValueCondition`
   - **Impact:** All configs using CompareStringFieldCondition must update component kind
   - **Reason:** Rename for consistency and clarity

2. **Expanded Operators:** New operators added
   - **Impact:** CompareStringFieldCondition operators (`=`, `!=`, `>`, `>=`, `<`, `<=`) replaced with (`=`, `!=`, `contains`, `does-not-contain`, `starts-with`)
   - **Reason:** String-specific operators more intuitive than lexicographic comparison

### Migration Steps

**Manual:**

Before (CompareStringFieldCondition):
```yaml
components:
  - name: check-environment
    kind: CompareStringFieldCondition
    properties:
      - name: Fields
        value: [environment]
      - name: Operator
        value: '='
      - name: Value
        value: production
```

After (StringValueCondition):
```yaml
components:
  - name: check-environment
    kind: StringValueCondition
    properties:
      - name: Fields
        value: [environment]
      - name: Operator
        value: '='
      - name: Value
        value: production
```

### Operator Migration

| CompareStringFieldCondition | StringValueCondition | Notes |
|----------------------------|---------------------|-------|
| `=` | `=` | Direct match |
| `!=` | `!=` | Direct match |
| `>`, `>=`, `<`, `<=` | N/A | Use NumericValueCondition for lexicographic comparison |

For lexicographic string comparisons (`>`, `>=`, `<`, `<=`), use [NumericValueCondition](../numeric_value_condition/README.md) if comparing numeric strings, or re-evaluate the condition logic.

### Backward Compatibility

None. This is a complete replacement requiring config updates.

### Additional Capabilities

StringValueCondition adds:
- `contains` - Check if field contains substring
- `does-not-contain` - Check if field doesn't contain substring
- `starts-with` - Check if field starts with prefix
- Smart scope selection based on operator
