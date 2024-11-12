package yaml

import (
	"fmt"
)

func AsString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func AsInt(v any) int {
	if v == nil {
		return 0
	}
	switch v := v.(type) {
	case bool:
		if v {
			return 1
		}
		return 0
	case int:
		return v
	case float64:
		return int(v)
	case string:
		var i int
		fmt.Sscanf(v, "%d", &i)
		return i
	default:
		return 0
	}
}

func AsFloat(v any) float64 {
	if v == nil {
		return 0
	}
	switch v := v.(type) {
	case bool:
		if v {
			return 1
		}
		return 0
	case int:
		return float64(v)
	case float64:
		return v
	case string:
		var f float64
		fmt.Sscanf(v, "%f", &f)
		return f
	default:
		return 0
	}
}
