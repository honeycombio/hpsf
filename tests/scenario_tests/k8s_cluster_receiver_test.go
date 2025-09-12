package hpsftests

import (
	"fmt"
	"testing"

	collectorprovider "github.com/honeycombio/hpsf/tests/providers/collector"
	hpsfprovider "github.com/honeycombio/hpsf/tests/providers/hpsf"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestK8sClusterReceiver(t *testing.T) {
	_, collectorConfig, _ := hpsfprovider.GetParsedConfigsFromFile(t, "testdata/k8s_cluster_receiver.yaml")

	// Verify the metrics pipeline exists and has the correct components
	metricsPipelines := collectorprovider.GetPipelinesByType(collectorConfig, "metrics")
	assert.Len(t, metricsPipelines, 1, "Expected 1 metrics pipeline, got %v", metricsPipelines)

	receivers, _, _, getResult := collectorprovider.GetPipelineConfig(collectorConfig, metricsPipelines[0].String())
	require.True(t, getResult.Found, "Expected metrics pipeline to be found")

	// Check pipeline components
	assert.Len(t, receivers, 1, "Expected 1 receiver")
	assert.Contains(t, receivers, "k8s_cluster/k8s_cluster_in", "Expected OTel receiver")

	config, result := collectorprovider.GetReceiverConfig[k8sclusterreceiver.Config](collectorConfig, "k8s_cluster/k8s_cluster_in")
	require.True(t, result.Found, "Expected receiver config to be found")
	assert.NotNil(t, config, "Expected receiver config to be non-nil")

	// Verify some key config values
	assert.Equal(t, "serviceAccount", fmt.Sprintf("%v", config.AuthType))
	assert.Equal(t, "k8s_leader_elector", fmt.Sprintf("%v", config.K8sLeaderElector))
	assert.NotNil(t, config.ResourceAttributes, "Expected resource attributes to be non-nil")

	// Verify resource attribute was enabled
	assert.Equal(t, true, config.ResourceAttributes.K8sContainerStatusLastTerminatedReason.Enabled)
}
