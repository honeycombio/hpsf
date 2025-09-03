package translator

import "testing"

func TestSubstitutor_DoSubstitutions(t *testing.T) {
	// use the default priorities
	tr := NewSubstitutor()
	tr.AddSubstitution("team", "apikey", "abc123")
	tr.AddSubstitution("team", "name", "myteam")
	tr.AddSubstitution("installation", "apikey", "def456")
	tr.AddSubstitution("installation", "name", "myinstall")
	tr.AddSubstitution("cluster", "apikey", "ghi789")
	tr.AddSubstitution("cluster", "name", "mycluster")
	tr.AddSubstitution("cluster", "foo.bar", "baz")

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"none", "no substitutions", "no substitutions"},
		{"one with context", "there->${team.name}<-there", "there->myteam<-there"},
		{"one without context", "substitute '${name}'", "substitute 'myteam'"},
		{"two with context", "a ${team.name} b ${cluster.apikey} c", "a myteam b ghi789 c"},
		{"two without context", "a ${name} b ${apikey} c", "a myteam b abc123 c"},
		{"two with different contexts", "a ${team.name} b ${cluster.name} c", "a myteam b mycluster c"},
		{"context not found", "a ${foo.name} b", "a ${foo.name} b"},
		{"context found, varname not found", "a ${team.foo} b", "a ${team.foo} b"},
		{"varname with embedded dot", "a ${foo.bar} b", "a baz b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tr.DoSubstitutions(tt.input); got != tt.want {
				t.Errorf("Substitutor.DoSubstitutions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSubstitutor_SubsWithPriorities(t *testing.T) {
	// set up priorities in reverse
	tr := NewSubstitutor()
	tr.AddSubstitution("cluster", "apikey", "ghi789")
	tr.AddSubstitution("cluster", "name", "mycluster")
	tr.AddSubstitution("cluster", "foo.bar", "baz")
	tr.AddSubstitution("installation", "apikey", "def456")
	tr.AddSubstitution("installation", "name", "myinstall")
	tr.AddSubstitution("team", "apikey", "abc123")
	tr.AddSubstitution("team", "name", "myteam")
	// now set explicit priorities
	tr.SetPriority("team", 3)
	tr.SetPriority("installation", 2)
	tr.SetPriority("cluster", 1)

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"none", "no substitutions", "no substitutions"},
		{"one with context", "there->${team.name}<-there", "there->myteam<-there"},
		{"one without context", "substitute '${name}'", "substitute 'myteam'"},
		{"two with context", "a ${team.name} b ${cluster.apikey} c", "a myteam b ghi789 c"},
		{"two without context", "a ${name} b ${apikey} c", "a myteam b abc123 c"},
		{"two with different contexts", "a ${team.name} b ${cluster.name} c", "a myteam b mycluster c"},
		{"context not found", "a ${foo.name} b", "a ${foo.name} b"},
		{"context found, varname not found", "a ${team.foo} b", "a ${team.foo} b"},
		{"varname with embedded dot", "a ${foo.bar} b", "a baz b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tr.DoSubstitutions(tt.input); got != tt.want {
				t.Errorf("Substitutor.DoSubstitutions() = %v, want %v", got, tt.want)
			}
		})
	}
}
