package internal

import (
	"testing"
)

func TestLinuxFilePermissions_String(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"rw-r--r--", "rw-r--r--"},
		{"rwxr-xr-x", "rwxr-xr-x"},
		{"rwsr-xr-x", "rwsr-xr-x"},
		{"rwxrwxrwx", "rwxrwxrwx"},
		{"--S--S--T", "--S--S--T"}, // Test setuid/setgid/sticky without exec
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := MustParse(tt.input)
			if got := p.String(); got != tt.want {
				t.Errorf("MustParse(%q).String() = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
