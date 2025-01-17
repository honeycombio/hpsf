package translator

import (
	"strings"
	"testing"

	"github.com/honeycombio/hpsf/pkg/config"
	"github.com/honeycombio/hpsf/pkg/hpsf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yamlv3 "gopkg.in/yaml.v3"
)

func TestGenerateConfig(t *testing.T) {
	const inputData = `components:
  - name: RefineryGRPC_2
    kind: RefineryGRPC
    ports:
      - name: TraceOut
        direction: output
        type: Honeycomb
    properties:
      - name: Port
        value: 4317
        type: number
  - name: otlp_in
    kind: OTelReceiver
    properties:
      - name: GRPCPort
        value: 9922
      - name: HTTPPort
        value: 1234
  - name: otlp_out
    kind: OTelGRPCExporter
    properties:
      - name: Host
        value: myhost.com
      - name: Port
        value: 1234
connections:
  - source:
      component: otlp_in
      port: Traces
      type: OTelTraces
    destination:
      component: otlp_out
      port: Traces
      type: OTelTraces`

	const expectedConfig = `receivers:
    otlp/otlp_in:
        protocols:
            grpc:
                endpoint: ${STRAWS_COLLECTOR_POD_IP}:9922
            http:
                endpoint: ${STRAWS_COLLECTOR_POD_IP}:1234
exporters:
    otlp/otlp_out:
        protocols:
            grpc:
                endpoint: myhost.com:1234
service:
    pipelines:
        traces:
            receivers: [otlp/otlp_in]
            processors: []
            exporters: [otlp/otlp_out]
`

	var hpsf *hpsf.HPSF
	dec := yamlv3.NewDecoder(strings.NewReader(inputData))
	err := dec.Decode(&hpsf)
	require.NoError(t, err)

	tlater, err := NewTranslator()
	require.NoError(t, err)

	cfg, err := tlater.GenerateConfig(hpsf, config.CollectorConfigType, nil)
	require.NoError(t, err)

	got, err := cfg.RenderYAML()
	require.NoError(t, err)

	assert.Equal(t, expectedConfig, string(got))
}
