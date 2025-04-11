# Creating components

## Anatomy of a component

```yaml
# kind is the unique type of the component;
# there can be only one component with any given kind
kind: OTelDebugExporter
# name is what the user calls the component.
# the name here is used to fill in the default name (a number will be appended)
name: OTel Debug Exporter
# style is used to control UI rendering
# supported values today are receiver, processor, exporter, sampler
style: exporter
# logo is used to define the logo used for receivers and exporters; no need to specify if not needed.
# the valid logos are listed in hound, in
# cmd/poodle/javascript/pipelines/ConfigurationVisualEditor/NodeComponentLogo.tsx
# if we need other logos for specialized components they can be added.
logo: opentelemetry
# type is 'base', 'meta', or 'template'
type: base
# status: alpha, stable, and archived are public
# use development if you don't want people to see it without a feature flag
status: alpha
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
tags:
  - category:exporter
  - category:debug
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
        value: "{{ firstNonZero .HProps.Verbosity .User.Verbosity .Props.Verbosity.Default }}"
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
compose differently from the standard. Today, `collector_config` always uses `collector` and the others always use `dottedConfig`. This could change
when composing samplers.

### meta

Each format can do different things with the meta section. For collectors,
`meta` contains:

- `componentSection` - exporters, receivers, processors etc
- `signalTypes` - array of which signal types this component handles
- `collectorComponentName` is the name by which the underlying collector component is known

### data

All the templates have a `data` section that is run through Go's text/templates
to convert them to configurations. For now, `data` is an array of elements, each of which supports 3 fields:

- `key` - the name of the yaml key under which this value will be stored. For non-collectors, this is a "dotted" key (meaning dots separate multiple levels) in YAML. For collectors, it's complicated. Read the code.
- `value` - the value of the key that should end up in the config
- `suppress_if` - a value that evaluates to nonzero if the entire key/value pair should be omitted. For example, we use this to output the `Insecure` flag only when it's true.
