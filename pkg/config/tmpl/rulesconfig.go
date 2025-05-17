package tmpl

import y "gopkg.in/yaml.v3"

type RulesConfig struct {
	Version int
	Envs    []EnvConfig
}

type EnvConfig struct {
	Name       string
	ConfigData DottedConfig
}

func NewRulesConfig() *RulesConfig {
	return &RulesConfig{
		Envs: []EnvConfig{},
	}
}

func (rc *RulesConfig) RenderToMap(m map[string]any) map[string]any {
	if m == nil {
		m = make(map[string]any)
	}
	m["RulesVersion"] = rc.Version
	for _, env := range rc.Envs {
		m = env.ConfigData.RenderToMap(m)
	}
	return m
}

func (rc *RulesConfig) RenderYAML() ([]byte, error) {
	m := rc.RenderToMap(nil)
	data, err := y.Marshal(m)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (rc *RulesConfig) Merge(other TemplateConfig) TemplateConfig {
	otherRC, ok := other.(*RulesConfig)
	if !ok {
		// if the other TemplateConfig is not a RulesConfig, we can't merge it
		return rc
	}
	for _, otherEnv := range otherRC.Envs {
		found := false
		for i, env := range rc.Envs {
			if env.Name == otherEnv.Name {
				rc.Envs[i].ConfigData = rc.Envs[i].ConfigData.Merge(otherEnv.ConfigData).(DottedConfig)
				found = true
				break
			}
		}
		if !found {
			rc.Envs = append(rc.Envs, otherEnv)
		}
	}
	return rc
}
