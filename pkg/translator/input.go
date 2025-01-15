package translator

import (
	"fmt"

	"github.com/honeycombio/hpsf/pkg/config"
	"github.com/honeycombio/hpsf/pkg/hpsf"
)

func NewInput(component hpsf.Component) (config.Component, error) {
	switch component.Kind {
	case "RefineryGRPC", "RefineryHTTP":
		return &config.RefineryInputComponent{Component: component}, nil
	case "TraceHTTP":
		return config.NewNullComponent(), nil
	case "LogHTTP":
		return config.NewNullComponent(), nil
	default:
		return nil, fmt.Errorf("unknown component kind: %s", component.Kind)
	}
}
