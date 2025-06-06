package tmpl

import (
	"reflect"
	"testing"
)

func TestDottedConfig_Render(t *testing.T) {
	tests := []struct {
		name string
		dc   DottedConfig
		want map[string]any
	}{
		{"0", DottedConfig{}, map[string]any{}},
		{"nodot", DottedConfig{"a": 1}, map[string]any{"a": 1}},
		{"1", DottedConfig{"a.b.c": 1}, map[string]any{"a": map[string]any{"b": map[string]any{"c": 1}}}},
		{"2", DottedConfig{"a.b.c": 1, "a.b.d": 2}, map[string]any{"a": map[string]any{"b": map[string]any{"c": 1, "d": 2}}}},
		{"3", DottedConfig{"a.b.c": 1, "a.b.d": 2, "a.e": 3}, map[string]any{"a": map[string]any{"b": map[string]any{"c": 1, "d": 2}, "e": 3}}},
		{"4", DottedConfig{"a.b.c": 1, "a.b.d": 2, "a.e": 3, "f": 4}, map[string]any{"a": map[string]any{"b": map[string]any{"c": 1, "d": 2}, "e": 3}, "f": 4}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.dc.RenderToMap(nil); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DottedConfig.Render() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeDottedConfig(t *testing.T) {
	type args struct {
		m map[string]any
	}
	tests := []struct {
		name string
		args args
		want DottedConfig
	}{
		{"0", args{map[string]any{}}, DottedConfig{}},
		{"1", args{map[string]any{"a": 1}}, DottedConfig{"a": 1}},
		{"2", args{map[string]any{"a": map[string]any{"b": 1}}}, DottedConfig{"a.b": 1}},
		{"3", args{map[string]any{"a": map[string]any{"b": 1, "c": 2}}}, DottedConfig{"a.b": 1, "a.c": 2}},
		{"4", args{map[string]any{"a": map[string]any{"b": 1, "c": 2}, "d": 3}}, DottedConfig{"a.b": 1, "a.c": 2, "d": 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDottedConfig(tt.args.m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeDottedConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func clone(m map[string]any) map[string]any {
	cl := make(map[string]any)
	for k, v := range m {
		switch v := v.(type) {
		case map[string]any:
			cl[k] = clone(v)
		default:
			cl[k] = v
		}
	}
	return cl
}

func TestDottedConfig_Compose(t *testing.T) {
	baseMap := map[string]any{"a": map[string]any{"b": map[string]any{"c": 1}}}
	tests := []struct {
		name string
		dc   DottedConfig
		want map[string]any
	}{
		{"1", DottedConfig{"b": 1}, map[string]any{"a": map[string]any{"b": map[string]any{"c": 1}}, "b": 1}},
		{"2", DottedConfig{"a.b.c": 2}, map[string]any{"a": map[string]any{"b": []map[string]any{{"c": 1}, {"c": 2}}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := clone(baseMap)
			if got := tt.dc.RenderToMap(m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DottedConfig.Render() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_processIndices(t *testing.T) {
	tests := []struct {
		name string
		in   map[string]any
		want map[string]any
	}{
		{"0", map[string]any{}, map[string]any{}},
		{"1", map[string]any{"a[0]": map[string]any{"a": "b"}}, map[string]any{"a": []any{map[string]any{"a": "b"}}}},
		{"2", map[string]any{"a[0]": "item1", "a[1]": "item2"}, map[string]any{"a": []any{"item1", "item2"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := processIndices(tt.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processIndices() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
