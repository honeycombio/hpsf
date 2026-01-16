# Migration Guide: AttributeJSONParsingProcessor

## AttributeJSONParsingProcessor → LogAttrJSONParsingProcessor & SpanAttrJSONParsingProcessor

**Status:** AttributeJSONParsingProcessor deprecated 2026-01-15
**Released:** 2026-01-15
**Replacements:**
- [LogAttrJSONParsingProcessor](../log_attr_json_parsing_processor/README.md) for log signals
- [SpanAttrJSONParsingProcessor](../span_attr_json_parsing_processor/README.md) for trace signals

### Breaking Changes

1. **Component Split by Signal:** Single component → Two signal-specific components
   - **Impact:** Must use correct component based on signal type
   - **Reason:** Clearer separation of concerns and simpler configuration

2. **Property Removal:** `Signal` property removed
   - **Impact:** Component selection determines signal type
   - **Reason:** No longer needed with signal-specific components

3. **Property Unchanged:** `Attribute` property same in both replacements
   - **Impact:** No changes needed to attribute name
   - **Reason:** Same functionality, just split by signal

### Migration Steps

**For Log Signals:**

Before (AttributeJSONParsingProcessor):
```yaml
components:
  - name: parse-json
    kind: AttributeJSONParsingProcessor
    properties:
      - name: Signal
        value: log
      - name: Attribute
        value: payload
```

After (LogAttrJSONParsingProcessor):
```yaml
components:
  - name: parse-json
    kind: LogAttrJSONParsingProcessor
    properties:
      - name: Attribute
        value: payload
```

**For Span/Trace Signals:**

Before (AttributeJSONParsingProcessor):
```yaml
components:
  - name: parse-json
    kind: AttributeJSONParsingProcessor
    properties:
      - name: Signal
        value: span
      - name: Attribute
        value: metadata
```

After (SpanAttrJSONParsingProcessor):
```yaml
components:
  - name: parse-json
    kind: SpanAttrJSONParsingProcessor
    properties:
      - name: Attribute
        value: metadata
```

### Migration Table

| AttributeJSONParsingProcessor | Replacement |
|------------------------------|-------------|
| Signal: `log` | LogAttrJSONParsingProcessor |
| Signal: `span` | SpanAttrJSONParsingProcessor |

### Backward Compatibility

None. This is a complete replacement requiring config updates.

### Additional Capabilities

The new signal-specific processors:
- Simpler configuration (no signal selection needed)
- Clearer intent and validation
- Same JSON parsing functionality
- Same error handling behavior
