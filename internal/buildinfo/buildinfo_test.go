package buildinfo

import "testing"

func TestCurrent(t *testing.T) {
	got := Current()
	if got.Version == "" || got.Commit == "" || got.Date == "" {
		t.Fatalf("Current() returned an empty field: %#v", got)
	}
}
