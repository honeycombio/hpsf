# Migration Guide: FieldContainsCondition

## FieldContainsCondition → StringValueCondition

**Status:** FieldContainsCondition deprecated 2026-01-15
**Released:** 2026-01-15
**Replacement:** [StringValueCondition](../string_value_condition/README.md)

### Breaking Changes

1. **Component Kind Change:** `FieldContainsCondition` → `StringValueCondition`
   - **Impact:** All configs using FieldContainsCondition must update component kind
   - **Reason:** Consolidate string comparison operations into single component

2. **Property Rename:** `Field` → `Fields` (now array)
   - **Impact:** Single field must be wrapped in array
   - **Reason:** Support checking multiple fields with same condition

3. **Property Rename:** `Substring` → `Value`
   - **Impact:** All configs must rename property
   - **Reason:** Generalize for all comparison types

4. **New Property:** `Operator` (set to `contains`)
   - **Impact:** Must explicitly set operator to `contains`
   - **Reason:** StringValueCondition supports multiple operators (=, !=, contains, does-not-contain, starts-with)

### Migration Steps

**Manual:**

Before (FieldContainsCondition):
```yaml
components:
  - name: check-substring
    kind: FieldContainsCondition
    properties:
      - name: Field
        value: error.message
      - name: Substring
        value: timeout
```

After (StringValueCondition):
```yaml
components:
  - name: check-substring
    kind: StringValueCondition
    properties:
      - name: Fields
        value: [error.message]
      - name: Operator
        value: contains
      - name: Value
        value: timeout
```

### Backward Compatibility

None. This is a complete replacement requiring config updates.

### Additional Capabilities

StringValueCondition adds:
- Support for multiple fields in single condition
- Additional operators: `=`, `!=`, `does-not-contain`, `starts-with`
- Smart scope selection based on operator
