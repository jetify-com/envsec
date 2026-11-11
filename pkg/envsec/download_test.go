package envsec

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteSecretFilePermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")

	// An existing world-readable file should be tightened.
	if err := os.WriteFile(path, []byte("OLD=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writeSecretFile(path, []byte("NEW=1\n")); err != nil {
		t.Fatalf("writeSecretFile() error = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("perm = %o, want 600", perm)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "NEW=1\n" {
		t.Errorf("contents = %q, want %q", got, "NEW=1\n")
	}
}

func TestValidateFormat(t *testing.T) {
	for _, f := range []string{"", "json", "dotenv"} {
		if err := ValidateFormat(f); err != nil {
			t.Errorf("ValidateFormat(%q) = %v, want nil", f, err)
		}
	}
	if err := ValidateFormat("yaml"); err == nil {
		t.Error("ValidateFormat(\"yaml\") = nil, want error")
	}
}
