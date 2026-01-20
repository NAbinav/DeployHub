package utils

import (
	"testing"
)

func TestParseENVString(t *testing.T) {
	envStr := "FOO=bar\nBAZ=qux\nAPI_KEY=secret123\n"
	got := ParseENVString(envStr)
	if len(got) != 3 {
		t.Errorf("expected 3 entries, got %d", len(got))
	}
	if got["FOO"] != "bar" {
		t.Errorf("expected FOO=bar, got FOO=%s", got["FOO"])
	}
	if got["BAZ"] != "qux" {
		t.Errorf("expected BAZ=qux, got BAZ=%s", got["BAZ"])
	}
	if got["API_KEY"] != "secret123" {
		t.Errorf("expected API_KEY=secret123, got API_KEY=%s", got["API_KEY"])
	}
}
