package envsec

import "testing"

func TestEnsureValidNames(t *testing.T) {
	valid := []string{"A", "_A", "abc_123", "FOO_BAR"}
	for _, name := range valid {
		if err := ensureValidNames([]string{name}); err != nil {
			t.Errorf("ensureValidNames(%q) = %v, want nil", name, err)
		}
	}

	invalid := []string{
		"",
		"1A",
		"A/../../other-project/prod/KEY",
		"PATH=/evil:/bin",
		"A B",
		"A-B",
		"A\n",
		"jetpack_foo",
	}
	for _, name := range invalid {
		if err := ensureValidNames([]string{name}); err == nil {
			t.Errorf("ensureValidNames(%q) = nil, want error", name)
		}
	}
}
