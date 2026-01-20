package utils

import "testing"

func TestENVStringRoundTrip(t *testing.T) {
	original := map[string]string{
		"FOO": "bar",
		"BAZ": "qux",
		"KEY": "value=with=equals",
	}

	// Convert to string and back
	str := ENVString(original)
	got := ParseENVString(str)

	// Check all keys exist and values match
	for key, val := range original {
		if got[key] != val {
			t.Errorf("roundtrip failed: %s expected %s, got %s", key, val, got[key])
		}
	}

	if len(got) != len(original) {
		t.Errorf("roundtrip failed: expected %d entries, got %d", len(original), len(got))
	}
}
