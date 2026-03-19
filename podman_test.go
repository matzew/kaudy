package kaudy

import (
	"strings"
	"testing"
)

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"MyProject", "myproject"},
		{"my-project", "my-project"},
		{"My Project!", "my-project-"},
		{"UPPER_CASE", "upper-case"},
		{"a1b2c3", "a1b2c3"},
	}
	for _, tt := range tests {
		got := sanitizeName(tt.input)
		if got != tt.want {
			t.Errorf("sanitizeName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestShelljoin(t *testing.T) {
	got := shelljoin([]string{"run", "--rm", "-it", "hello world"})
	if !strings.Contains(got, `"hello world"`) {
		t.Errorf("shelljoin did not quote space-containing arg: %s", got)
	}
	if strings.Contains(got, `"run"`) {
		t.Errorf("shelljoin should not quote simple args: %s", got)
	}
}
