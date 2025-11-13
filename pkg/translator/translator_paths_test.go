package translator

import (
	"fmt"
	"testing"

	"github.com/honeycombio/hpsf/pkg/config"
	"github.com/honeycombio/hpsf/pkg/hpsf"
	"github.com/honeycombio/hpsf/pkg/hpsftypes"
)

// TestOrderPathsPortIndex ensures that orderPaths sorts by port index when present.
func TestOrderPathsPortIndex(t *testing.T) {
	start := config.TemplateComponent{
		Kind: "StartSamplingFake",
		Name: "StartSampler",
		Ports: []config.TemplatePort{
			{Name: "outB", Direction: "output", Index: 2},
			{Name: "outA", Direction: "output", Index: 1},
			{Name: "outC", Direction: "output"}, // unspecified index
		},
		Style: "startsampling",
	}
	down1 := config.TemplateComponent{Kind: "SamplerA", Name: "SamplerA"}
	down2 := config.TemplateComponent{Kind: "SamplerB", Name: "SamplerB"}
	down3 := config.TemplateComponent{Kind: "SamplerC", Name: "SamplerC"}

	comps := NewOrderedComponentMap()
	comps.Set("StartSampler", &start)
	comps.Set("SamplerA", &down1)
	comps.Set("SamplerB", &down2)
	comps.Set("SamplerC", &down3)

	connB := &hpsf.Connection{Source: hpsf.ConnectionPort{Component: "StartSampler", PortName: "outB"}, Destination: hpsf.ConnectionPort{Component: "SamplerB"}}
	connA := &hpsf.Connection{Source: hpsf.ConnectionPort{Component: "StartSampler", PortName: "outA"}, Destination: hpsf.ConnectionPort{Component: "SamplerA"}}
	connC := &hpsf.Connection{Source: hpsf.ConnectionPort{Component: "StartSampler", PortName: "outC"}, Destination: hpsf.ConnectionPort{Component: "SamplerC"}}

	paths := []hpsf.PathWithConnections{
		{ConnType: hpsf.CTYPE_TRACES, Path: []*hpsf.Component{{Name: "StartSampler"}, {Name: "SamplerB"}}, Connections: []*hpsf.Connection{connB}},
		{ConnType: hpsf.CTYPE_TRACES, Path: []*hpsf.Component{{Name: "StartSampler"}, {Name: "SamplerA"}}, Connections: []*hpsf.Connection{connA}},
		{ConnType: hpsf.CTYPE_TRACES, Path: []*hpsf.Component{{Name: "StartSampler"}, {Name: "SamplerC"}}, Connections: []*hpsf.Connection{connC}},
	}

	orderPaths(paths, comps)

	if paths[0].Connections[0].Source.PortName != "outA" ||
		paths[1].Connections[0].Source.PortName != "outB" ||
		paths[2].Connections[0].Source.PortName != "outC" {
		for i, p := range paths {
			if len(p.Connections) == 0 {
				continue
			}
			t.Errorf("unexpected order at position %d: %s", i, p.Connections[0].Source.PortName)
		}
	}
}

// TestOrderPathsIndexBeforeComponent verifies index ordering precedes component name ordering.
func TestOrderPathsIndexBeforeComponent(t *testing.T) {
	startA := config.TemplateComponent{ // index 2
		Kind: "StartA", Name: "StartA",
		Ports: []config.TemplatePort{{Name: "out1", Direction: "output", Index: 2}},
	}
	startZ := config.TemplateComponent{ // index 1
		Kind: "StartZ", Name: "StartZ",
		Ports: []config.TemplatePort{{Name: "out1", Direction: "output", Index: 1}},
	}
	down1 := config.TemplateComponent{Kind: "Down1", Name: "Down1"}
	down2 := config.TemplateComponent{Kind: "Down2", Name: "Down2"}

	comps := NewOrderedComponentMap()
	comps.Set("StartA", &startA)
	comps.Set("StartZ", &startZ)
	comps.Set("Down1", &down1)
	comps.Set("Down2", &down2)

	connA := &hpsf.Connection{Source: hpsf.ConnectionPort{Component: "StartA", PortName: "out1"}, Destination: hpsf.ConnectionPort{Component: "Down1"}}
	connZ := &hpsf.Connection{Source: hpsf.ConnectionPort{Component: "StartZ", PortName: "out1"}, Destination: hpsf.ConnectionPort{Component: "Down2"}}

	paths := []hpsf.PathWithConnections{
		{ConnType: hpsf.CTYPE_TRACES, Path: []*hpsf.Component{{Name: "StartZ"}, {Name: "Down2"}}, Connections: []*hpsf.Connection{connZ}}, // index 1
		{ConnType: hpsf.CTYPE_TRACES, Path: []*hpsf.Component{{Name: "StartA"}, {Name: "Down1"}}, Connections: []*hpsf.Connection{connA}}, // index 2
	}

	orderPaths(paths, comps)

	if paths[0].Connections[0].Source.Component != "StartZ" { // lower index should come first
		for i, p := range paths {
			if len(p.Connections) == 0 {
				continue
			}
			t.Errorf("unexpected order at %d: %s", i, p.Connections[0].Source.Component)
		}
	}
}

// TestSamplingSequencerRuleOrdering ensures that the path ordering logic yields the
// rule output ports in strict numeric order (Rule 1 .. Rule 12) when each rule
// output of the SamplingSequencer fans out into a distinct SampleData path.
func TestSamplingSequencerRuleOrdering(t *testing.T) {
	hpsfYAML := `kind: Test
components:
  - name: Receive OTel
    kind: OTelReceiver
    version: v0.1.0
  - name: Start Sampling
    kind: SamplingSequencer
    version: v0.1.0
  - name: Send to Honeycomb
    kind: HoneycombExporter
    version: v0.1.0
  - name: Keep All
    kind: KeepAllSampler
    version: v0.1.0
  - name: Check Duration_1
    kind: LongDurationCondition
    version: v0.1.0
  - name: Drop_1
    kind: Dropper
    version: v0.1.0
  - name: Check Duration_2
    kind: LongDurationCondition
    version: v0.1.0
  - name: Drop_2
    kind: Dropper
    version: v0.1.0
  - name: Check Duration_3
    kind: LongDurationCondition
    version: v0.1.0
  - name: Drop_3
    kind: Dropper
    version: v0.1.0
  - name: Check Duration_4
    kind: LongDurationCondition
    version: v0.1.0
  - name: Drop_4
    kind: Dropper
    version: v0.1.0
  - name: Check Duration_5
    kind: LongDurationCondition
    version: v0.1.0
  - name: Drop_5
    kind: Dropper
    version: v0.1.0
  - name: Check Duration_6
    kind: LongDurationCondition
    version: v0.1.0
  - name: Drop_6
    kind: Dropper
    version: v0.1.0
  - name: Check Duration_7
    kind: LongDurationCondition
    version: v0.1.0
  - name: Drop_7
    kind: Dropper
    version: v0.1.0
  - name: Check Duration_8
    kind: LongDurationCondition
    version: v0.1.0
  - name: Drop_8
    kind: Dropper
    version: v0.1.0
  - name: Check Duration_9
    kind: LongDurationCondition
    version: v0.1.0
  - name: Drop_9
    kind: Dropper
    version: v0.1.0
  - name: Check Duration_10
    kind: LongDurationCondition
    version: v0.1.0
  - name: Drop_10
    kind: Dropper
    version: v0.1.0
  - name: Check Duration_11
    kind: LongDurationCondition
    version: v0.1.0
  - name: Drop_11
    kind: Dropper
    version: v0.1.0
  - name: Check Duration_12
    kind: LongDurationCondition
    version: v0.1.0
  - name: Drop_12
    kind: Dropper
    version: v0.1.0
  - name: Check Duration_13
    kind: LongDurationCondition
    version: v0.1.0
  - name: Drop_13
    kind: Dropper
    version: v0.1.0
  - name: Check Duration_14
    kind: LongDurationCondition
    version: v0.1.0
  - name: Drop_14
    kind: Dropper
    version: v0.1.0
  - name: Check Duration_15
    kind: LongDurationCondition
    version: v0.1.0
  - name: Drop_15
    kind: Dropper
    version: v0.1.0
  - name: Check Duration_16
    kind: LongDurationCondition
    version: v0.1.0
  - name: Drop_16
    kind: Dropper
    version: v0.1.0
  - name: Check Duration_17
    kind: LongDurationCondition
    version: v0.1.0
  - name: Drop_17
    kind: Dropper
    version: v0.1.0
  - name: Check Duration_18
    kind: LongDurationCondition
    version: v0.1.0
  - name: Drop_18
    kind: Dropper
    version: v0.1.0
  - name: Check Duration_19
    kind: LongDurationCondition
    version: v0.1.0
  - name: Drop_19
    kind: Dropper
    version: v0.1.0
  - name: Check Duration_20
    kind: LongDurationCondition
    version: v0.1.0
  - name: Drop_20
    kind: Dropper
    version: v0.1.0
  - name: Check Duration_21
    kind: LongDurationCondition
    version: v0.1.0
  - name: Drop_21
    kind: Dropper
    version: v0.1.0
  - name: Check Duration_22
    kind: LongDurationCondition
    version: v0.1.0
  - name: Drop_22
    kind: Dropper
    version: v0.1.0
  - name: Check Duration_23
    kind: LongDurationCondition
    version: v0.1.0
  - name: Drop_23
    kind: Dropper
    version: v0.1.0
  - name: Check Duration_24
    kind: LongDurationCondition
    version: v0.1.0
  - name: Drop_24
    kind: Dropper
    version: v0.1.0
connections:
  - source: { component: Receive OTel, port: Traces, type: OTelTraces }
    destination: { component: Start Sampling, port: Traces, type: OTelTraces }
  - source: { component: Receive OTel, port: Logs, type: OTelLogs }
    destination: { component: Start Sampling, port: Logs, type: OTelLogs }
  - source: { component: Start Sampling, port: Rule 1, type: SampleData }
    destination: { component: Check Duration_1, port: Match, type: SampleData }
  - source: { component: Check Duration_1, port: And, type: SampleData }
    destination: { component: Drop_1, port: Sample, type: SampleData }
  - source: { component: Start Sampling, port: Rule 2, type: SampleData }
    destination: { component: Check Duration_2, port: Match, type: SampleData }
  - source: { component: Check Duration_2, port: And, type: SampleData }
    destination: { component: Drop_2, port: Sample, type: SampleData }
  - source: { component: Start Sampling, port: Rule 3, type: SampleData }
    destination: { component: Check Duration_3, port: Match, type: SampleData }
  - source: { component: Check Duration_3, port: And, type: SampleData }
    destination: { component: Drop_3, port: Sample, type: SampleData }
  - source: { component: Start Sampling, port: Rule 4, type: SampleData }
    destination: { component: Check Duration_4, port: Match, type: SampleData }
  - source: { component: Check Duration_4, port: And, type: SampleData }
    destination: { component: Drop_4, port: Sample, type: SampleData }
  - source: { component: Start Sampling, port: Rule 5, type: SampleData }
    destination: { component: Check Duration_5, port: Match, type: SampleData }
  - source: { component: Check Duration_5, port: And, type: SampleData }
    destination: { component: Drop_5, port: Sample, type: SampleData }
  - source: { component: Start Sampling, port: Rule 6, type: SampleData }
    destination: { component: Check Duration_6, port: Match, type: SampleData }
  - source: { component: Check Duration_6, port: And, type: SampleData }
    destination: { component: Drop_6, port: Sample, type: SampleData }
  - source: { component: Start Sampling, port: Rule 7, type: SampleData }
    destination: { component: Check Duration_7, port: Match, type: SampleData }
  - source: { component: Check Duration_7, port: And, type: SampleData }
    destination: { component: Drop_7, port: Sample, type: SampleData }
  - source: { component: Start Sampling, port: Rule 8, type: SampleData }
    destination: { component: Check Duration_8, port: Match, type: SampleData }
  - source: { component: Check Duration_8, port: And, type: SampleData }
    destination: { component: Drop_8, port: Sample, type: SampleData }
  - source: { component: Start Sampling, port: Rule 9, type: SampleData }
    destination: { component: Check Duration_9, port: Match, type: SampleData }
  - source: { component: Check Duration_9, port: And, type: SampleData }
    destination: { component: Drop_9, port: Sample, type: SampleData }
  - source: { component: Start Sampling, port: Rule 10, type: SampleData }
    destination: { component: Check Duration_10, port: Match, type: SampleData }
  - source: { component: Check Duration_10, port: And, type: SampleData }
    destination: { component: Drop_10, port: Sample, type: SampleData }
  - source: { component: Start Sampling, port: Rule 11, type: SampleData }
    destination: { component: Check Duration_11, port: Match, type: SampleData }
  - source: { component: Check Duration_11, port: And, type: SampleData }
    destination: { component: Drop_11, port: Sample, type: SampleData }
  - source: { component: Start Sampling, port: Rule 12, type: SampleData }
    destination: { component: Check Duration_12, port: Match, type: SampleData }
  - source: { component: Check Duration_12, port: And, type: SampleData }
    destination: { component: Drop_12, port: Sample, type: SampleData }
  - source: { component: Start Sampling, port: Rule 13, type: SampleData }
    destination: { component: Check Duration_13, port: Match, type: SampleData }
  - source: { component: Check Duration_13, port: And, type: SampleData }
    destination: { component: Drop_13, port: Sample, type: SampleData }
  - source: { component: Start Sampling, port: Rule 14, type: SampleData }
    destination: { component: Check Duration_14, port: Match, type: SampleData }
  - source: { component: Check Duration_14, port: And, type: SampleData }
    destination: { component: Drop_14, port: Sample, type: SampleData }
  - source: { component: Start Sampling, port: Rule 15, type: SampleData }
    destination: { component: Check Duration_15, port: Match, type: SampleData }
  - source: { component: Check Duration_15, port: And, type: SampleData }
    destination: { component: Drop_15, port: Sample, type: SampleData }
  - source: { component: Start Sampling, port: Rule 16, type: SampleData }
    destination: { component: Check Duration_16, port: Match, type: SampleData }
  - source: { component: Check Duration_16, port: And, type: SampleData }
    destination: { component: Drop_16, port: Sample, type: SampleData }
  - source: { component: Start Sampling, port: Rule 17, type: SampleData }
    destination: { component: Check Duration_17, port: Match, type: SampleData }
  - source: { component: Check Duration_17, port: And, type: SampleData }
    destination: { component: Drop_17, port: Sample, type: SampleData }
  - source: { component: Start Sampling, port: Rule 18, type: SampleData }
    destination: { component: Check Duration_18, port: Match, type: SampleData }
  - source: { component: Check Duration_18, port: And, type: SampleData }
    destination: { component: Drop_18, port: Sample, type: SampleData }
  - source: { component: Start Sampling, port: Rule 19, type: SampleData }
    destination: { component: Check Duration_19, port: Match, type: SampleData }
  - source: { component: Check Duration_19, port: And, type: SampleData }
    destination: { component: Drop_19, port: Sample, type: SampleData }
  - source: { component: Start Sampling, port: Rule 20, type: SampleData }
    destination: { component: Check Duration_20, port: Match, type: SampleData }
  - source: { component: Check Duration_20, port: And, type: SampleData }
    destination: { component: Drop_20, port: Sample, type: SampleData }
  - source: { component: Start Sampling, port: Rule 21, type: SampleData }
    destination: { component: Check Duration_21, port: Match, type: SampleData }
  - source: { component: Check Duration_21, port: And, type: SampleData }
    destination: { component: Drop_21, port: Sample, type: SampleData }
  - source: { component: Start Sampling, port: Rule 22, type: SampleData }
    destination: { component: Check Duration_22, port: Match, type: SampleData }
  - source: { component: Check Duration_22, port: And, type: SampleData }
    destination: { component: Drop_22, port: Sample, type: SampleData }
  - source: { component: Start Sampling, port: Rule 23, type: SampleData }
    destination: { component: Check Duration_23, port: Match, type: SampleData }
  - source: { component: Check Duration_23, port: And, type: SampleData }
    destination: { component: Drop_23, port: Sample, type: SampleData }
  - source: { component: Start Sampling, port: Rule 24, type: SampleData }
    destination: { component: Check Duration_24, port: Match, type: SampleData }
  - source: { component: Check Duration_24, port: And, type: SampleData }
    destination: { component: Drop_24, port: Sample, type: SampleData }
  - source: { component: Start Sampling, port: Rule 25, type: SampleData }
    destination: { component: Keep All, port: Sample, type: SampleData }
  - source: { component: Keep All, port: Events, type: HoneycombEvents }
    destination: { component: Send to Honeycomb, port: Events, type: HoneycombEvents }
`

	h, err := hpsf.FromYAML(hpsfYAML)
	if err != nil {
		t.Fatalf("failed to parse HPSF yaml: %v", err)
	}

	tr := NewEmptyTranslator()
	if err := tr.LoadEmbeddedComponents(); err != nil {
		t.Fatalf("failed to load embedded components: %v", err)
	}

	_ = tr.ValidateConfig(&h)
	_, _ = tr.GenerateConfig(&h, hpsftypes.CollectorConfig, "latest", nil)

	paths := h.FindAllPaths(map[string]bool{})
	comps := NewOrderedComponentMap()
	for _, c := range h.Components {
		cc, err2 := tr.makeConfigComponent(c, hpsftypes.CollectorConfig, "latest")
		if err2 != nil {
			continue
		}
		comps.Set(c.GetSafeName(), cc)
	}
	orderPaths(paths, comps)

	rulePorts := make([]string, 0, 25)
	for _, p := range paths {
		if p.ConnType != hpsf.CTYPE_SAMPLE || len(p.Connections) == 0 {
			continue
		}
		first := p.Connections[0]
		if first.Source.Component == "Start Sampling" {
			rulePorts = append(rulePorts, first.Source.PortName)
		}
	}

	if len(rulePorts) != 25 {
		t.Fatalf("expected %d rule paths, got %d (%v)", 12, len(rulePorts), rulePorts)
	}
	for i, rp := range rulePorts {
		expected := fmt.Sprintf("Rule %d", i+1)
		if rp != expected {
			t.Fatalf("rule ordering mismatch at %d: got %s expected %s full=%v", i, rp, expected, rulePorts)
		}
	}
}
