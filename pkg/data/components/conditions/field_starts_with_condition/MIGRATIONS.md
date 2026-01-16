# Migration Guide: FieldStartsWithCondition

## FieldStartsWithCondition → StringValueCondition

**Status:** FieldStartsWithCondition deprecated 2026-01-15
**Released:** 2026-01-15
**Replacement:** [StringValueCondition](../string_value_condition/README.md)

### Breaking Changes

1. **Component Kind Change:** `FieldStartsWithCondition` → `StringValueCondition`
   - **Impact:** All configs using FieldStartsWithCondition must update component kind
   - **Reason:** Consolidate string comparison operations into single component

2. **Property Rename:** `Field` → `Fields` (now array)
   - **Impact:** Single field must be wrapped in array
   - **Reason:** Support checking multiple fields with same condition

3. **Property Rename:** `Prefix` → `Value`
   - **Impact:** All configs must rename property
   - **Reason:** Generalize for all comparison types

4. **New Property:** `Operator` (set to `starts-with`)
   - **Impact:** Must explicitly set operator to `starts-with`
   - **Reason:** StringValueCondition supports multiple operators (=, !=, contains, does-not-contain, starts-with)

### Migration Steps

**Manual:**

Before (FieldStartsWithCondition):
```yaml
components:
  - name: check-prefix
    kind: FieldStartsWithCondition
    properties:
      - name: Field
        value: http.route
      - name: Prefix
        value: /api/
```

After (StringValueCondition):
```yaml
components:
  - name: check-prefix
    kind: StringValueCondition
    properties:
      - name: Fields
        value: [http.route]
      - name: Operator
        value: starts-with
      - name: Value
        value: /api/
```

### Backward Compatibility

None. This is a complete replacement requiring config updates.

### Additional Capabilities

StringValueCondition adds:
- Support for multiple fields in single condition
- Additional operators: `=`, `!=`, `contains`, `does-not-contain`
- Smart scope selection based on operator
