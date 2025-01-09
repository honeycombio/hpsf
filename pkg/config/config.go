package config

// DefaultConfiguration is the default HPSF configuration that includes a
// simple Refinery configuration with a determinisic sampler.
const DefaultConfiguration = `
components:
  - name: DefaultDeterministicSampler
    kind: DeterministicSampler
    properties:
      - name: Environment
        value: __default__
        type: string
      - name: SampleRate
        value: 1
        type: number
`
