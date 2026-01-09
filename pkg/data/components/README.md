# Creating components

## Anatomy of a component

```yaml
# kind is the unique type of the component;
# there can be only one component with any given kind
# changing kind for a component in use is a BREAKING change and requires a major version bump
kind: OTelDebugExporter
# name is what the user calls the component.
# the name here is used to fill in the default name (a number will be appended)
name: OTel Debug Exporter
# style is used to control UI rendering
# supported values today are receiver, processor, exporter, sampler, condition
style: exporter
# logo is used to define the logo used for receivers and exporters; no need to specify if not needed.
# the valid logos are listed in hound, in
# cmd/poodle/javascript/pipelines/ConfigurationVisualEditor/NodeComponentLogo.tsx
# if we need other logos for specialized components they can be added.
logo: opentelemetry
# type is 'base', 'meta', or 'template'
type: base
# status lifecycle: development -> beta -> stable -> deprecated -> archived
# development: feature-flagged, open to changes
# beta: public, stable API, ready for users
# stable: production-ready, guaranteed stability
# deprecated: planned for removal
# archived: removed from active use
status: beta
# version should be bumped when the component is updated
version: v0.1.0
# summary is the short description, easily visible in the UI
# and the sidebar
summary: Sends pipeline signal traffic to stdout for debugging.
# description is longer and only shows up on demand
description: |
  Exports signal data from a pipeline to stdout. This is useful for debugging, but only if you
  have access to the stdout stream in your environment. This component is not intended for production use.
# tags are to help the user find and organize the component
# in the sidebar. follow the key:value format.
# There should be only one category tag (at least for now).
# Category values should be one of these: input, processor, startsampling, condition, sampler, output
# Service values (one of): collector, refinery
# Signal values (one of): OTelTraces, OTelMetrics, OTelLogs, HoneycombEvents, SampleData
tags:
  - category:output
  - service:collector
  - signal:OTelTraces
  - signal:OTelMetrics
  - signal:OTelLogs
# ports are the things that allow connections to other components
ports:
  # name is the name that shows up in the UI next to the handle
  # it's used in connections between components.
  # changing name across versions is a BREAKING change and requires a major version bump.
  - name: Traces
    # note that receivers have output ports and exporters have input ports.
    direction: input
    # be careful to specify the port types accurately.
    type: OTelTraces
  - name: Metrics
    direction: input
    type: OTelMetrics
  - name: Logs
    direction: input
    type: OTelLogs
# properties are the user-editable values for this component
properties:
  # the name of the property; this is used by the templates
  # so the name should be a valid Go identifier
  - name: Verbosity
    # summary shows up in the UI
    summary: The verbosity level of the debug output.
    # description is an on-demand longer description
    description: |
      The verbosity level of the debug output. Valid values are basic, normal, or detailed. The default is "basic".
    # type is the datatype of the value and partly controls the
    # property editor that will be used for this value
    # supported types: string, int, float, bool, stringarray
    type: string
    # subtype can further constrain the property editor;
    # in this case, a oneof() subtype will cause a dropdown
    # to be used instead of a text box
    subtype: oneof(basic, normal, detailed)
    # validations are constraints on the value, and should be
    # thought of as independent of the property editor.
    # this is because human-written code may not use the property editor.
    # permitted validations can be found in templateComponent.go
    validations:
      - noblanks
      - oneof(basic, normal, detailed)
    # default is the default value for the property
    default: basic
    # if advanced is true, this property shows up under "Advanced" and is hidden by default
    advanced: false
# templates control how this component is rendered in configurations
# there can be multiple entries in this array if the component
# can generate more than one template
# kind can be collector_config, refinery_config, or refinery_rules
# the template kind determines the rest of the fields and
# how they're interpreted. See below for more details
templates:
  - kind: collector_config
    name: otel_debug_exporter_collector
    format: collector
    meta:
      componentSection: exporters
      signalTypes: [traces, metrics, logs] # we'll generate a name for each pipeline if there's more than 1
      collectorComponentName: debug
    data:
      - key: "{{ .ComponentName }}.verbosity"
        value: "{{ .Values.Verbosity }}"
```

## The template section

As noted above, the template section is what does the work to generate
configurations from the data specified in each component.

### kind

At this writing, there are 3 kinds -- `collector_config`, `refinery_config`, and `refinery_rules`.

### name

This specifies a name for this template. It's currently not used for anything.
It's probably a good idea to make it unique in case we decide it's helpful somewhere.

### format

This is an escape hatch in case there are specialized components that need to
compose differently from the standard. Today,

- `collector_config` uses `collector`
- `refinery_config` uses `dottedConfig`
- `refinery_rules` uses `rules`
  when composing samplers.

### meta

Each format can do different things with the meta section.

For collectors, `meta` contains:

- `componentSection` - exporters, receivers, processors etc
- `signalTypes` - array of which signal types this component handles
- `collectorComponentName` is the name by which the underlying collector component is known

For refinery rules, `meta` contains:

- `env` - the environment for the rules (used for samplers, but not currently exposed to users)
- `sampler` - the kind of sampler being configured (the name used in Refinery configs, such as "DeterministicSampler")
- `condition` - set to `true` for condition components
- `scope` - optional field that can be set to "span" or "trace" to control condition evaluation scope. There is a ForceSpanScope processor that gives user control, but there are definitely cases where it makes more sense to force it in the component without making the user think about it.

These fields can use template variables (for example, `scope` often uses templating based on operator values).

### data

All the templates have a `data` section that is run through Go's text/templates
to convert them to configurations. Template values can use functions like `encodeAsInt`, `encodeAsArray`, etc.
For now, `data` is an array of elements, each of which supports 3 fields:

- `key` - the name of the yaml key under which this value will be stored. For non-collectors, this is a "dotted" key (meaning dots separate multiple levels) in YAML. For collectors, it's complicated. Read the code.
- `value` - the value of the key that should end up in the config
- `suppress_if` - a value that evaluates to nonzero if the entire key/value pair should be omitted. For example, we use this to output the `Insecure` flag only when it's true.

For refinery rules that are conditions, the data section contains structured key-value pairs that define the condition. These are the same keys used in normal Refinery rules.

```yaml
data:
  - key: Fields
    value: [http.status_code, http.response.status_code]
  - key: Operator
    value: ">="
  - key: Value
    value: "{{ .Values.Value | encodeAsInt }}"
  - key: Datatype
    value: int
```

These are the standards:

- `Fields` -- array of field names to check against (if there's only one, `Field` is supported without the array)
- `Operator` -- Refinery operator (=, !=, <, <=, >, >=, etc., as well as things like "contains" and "matches")
- `Value` -- the constant value being compared (often templated from properties, but sometimes hardcoded or unneeded, depending on the operator)
- `Datatype` -- the data type for comparison when it's appropriate to force it (string, int, float, bool). It's usually a good idea to specify this and not to leave it in the user's hands.

For both collector and refinery rules, the dottedconfig also supports fields
with a number in square brackets. If at any level, the key ends with a number in
square brackets (which indicates that it's an indexed value in a slice), then we
take the value of that key, determine its type T, and put it into a []T at the
same level, but with the new key being the portion of the name before the `[`
and `]`. The number in the brackets is the index of the slice.

Note that this does NOT apply for conditions, because conditions need to be in an array at the condition level instead of the field level. To generate multiple conditions in a single template, add .1, .2, etc to the names of the fields. Example:

```
    data:
      - key: Fields.1
        value: "{{ .Values.Fields | encodeAsArray }}"
      - key: Operator.1
        value: ">="
```

etc.
