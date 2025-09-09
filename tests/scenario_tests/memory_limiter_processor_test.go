package hpsftests

import (
	"os"
	"testing"

	"github.com/honeycombio/hpsf/pkg/hpsf"
	"github.com/stretchr/testify/require"
)

func TestMemoryLimiterProcessor(t *testing.T) {
	// Read the HPSF file directly and parse it
	file, err := os.ReadFile("testdata/memory_limiter_processor.yaml")
	require.NoError(t, err, "Failed to read HPSF file")

	h, err := hpsf.FromYAML(string(file))
	require.NoError(t, err, "HPSF should parse without errors")

	// Verify the HPSF components are correctly defined
	require.Len(t, h.Components, 3, "Expected 3 components: receiver, processor, exporter")

	// Find the memory limiter processor
	var memoryLimiterComponent *hpsf.Component
	for _, comp := range h.Components {
		if comp.Kind == "MemoryLimiterProcessor" {
			memoryLimiterComponent = comp
			break
		}
	}
	require.NotNil(t, memoryLimiterComponent, "Expected to find MemoryLimiterProcessor component")
	require.Equal(t, "memory_limiter", memoryLimiterComponent.Name)

	// Verify the memory limiter has the expected percentage-based properties
	require.Len(t, memoryLimiterComponent.Properties, 3, "Expected 3 properties for memory limiter (CheckInterval, LimitPercentage, SpikeLimitPercentage)")

	// Verify connections - memory limiter should be in the processing path
	require.Len(t, h.Connections, 6, "Expected 6 connections for 3 signal types")
}
