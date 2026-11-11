package envcli

import (
	"testing"

	"go.jetify.com/envsec/pkg/envsec"
)

// The --format flag's default must be a value that ValidateFormat accepts,
// otherwise a plain `envsec download .env` always fails.
func TestDownloadFormatDefaultIsValid(t *testing.T) {
	flag := DownloadCmd().Flags().Lookup("format")
	if flag == nil {
		t.Fatal("download command has no --format flag")
	}
	if err := envsec.ValidateFormat(flag.DefValue); err != nil {
		t.Errorf("default --format %q is rejected: %v", flag.DefValue, err)
	}
}
