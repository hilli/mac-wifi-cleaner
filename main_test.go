package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadWriteSSIDFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ssids.txt")

	original := []string{"HomeWifi", "CoffeeShop", "Office 5G"}
	if err := writeSSIDFile(path, original); err != nil {
		t.Fatalf("writeSSIDFile: %v", err)
	}

	got, err := readSSIDFile(path)
	if err != nil {
		t.Fatalf("readSSIDFile: %v", err)
	}

	if len(got) != len(original) {
		t.Fatalf("got %d SSIDs, want %d", len(got), len(original))
	}
	for i := range got {
		if got[i] != original[i] {
			t.Errorf("SSID[%d] = %q, want %q", i, got[i], original[i])
		}
	}
}

func TestReadSSIDFileComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ssids.txt")

	content := "# This is a comment\nAlpha\n\n# Another comment\nBravo\n  \n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := readSSIDFile(path)
	if err != nil {
		t.Fatalf("readSSIDFile: %v", err)
	}

	want := []string{"Alpha", "Bravo"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("SSID[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestReadSSIDFileEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.txt")
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := readSSIDFile(path)
	if err != nil {
		t.Fatalf("readSSIDFile: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}

func TestReadSSIDFileNotFound(t *testing.T) {
	_, err := readSSIDFile("/nonexistent/path/file.txt")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestToSet(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		check string
		want  bool
	}{
		{"present", []string{"a", "b", "c"}, "b", true},
		{"absent", []string{"a", "b", "c"}, "d", false},
		{"empty", []string{}, "a", false},
		{"duplicates", []string{"a", "a", "b"}, "a", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := toSet(tt.input)
			if got := s[tt.check]; got != tt.want {
				t.Errorf("toSet(%v)[%q] = %v, want %v", tt.input, tt.check, got, tt.want)
			}
		})
	}
}

func TestFlagValue(t *testing.T) {
	tests := []struct {
		name string
		args []string
		flag string
		want string
	}{
		{"present", []string{"-f", "file.txt", "--dry-run"}, "-f", "file.txt"},
		{"absent", []string{"--dry-run"}, "-f", ""},
		{"flag at end without value", []string{"--dry-run", "-f"}, "-f", ""},
		{"empty args", []string{}, "-o", ""},
		{"multiple flags", []string{"-o", "out.txt", "-f", "in.txt"}, "-f", "in.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := flagValue(tt.args, tt.flag)
			if got != tt.want {
				t.Errorf("flagValue(%v, %q) = %q, want %q", tt.args, tt.flag, got, tt.want)
			}
		})
	}
}

func TestHasFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
		flag string
		want bool
	}{
		{"present", []string{"-f", "file.txt", "--dry-run"}, "--dry-run", true},
		{"absent", []string{"-f", "file.txt"}, "--dry-run", false},
		{"empty", []string{}, "--dry-run", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasFlag(tt.args, tt.flag)
			if got != tt.want {
				t.Errorf("hasFlag(%v, %q) = %v, want %v", tt.args, tt.flag, got, tt.want)
			}
		})
	}
}
