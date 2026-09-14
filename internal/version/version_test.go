package version

import "testing"

func TestCurrent(t *testing.T) {
	got := Current()
	if got.Version == "" || got.Commit == "" || got.BuildDate == "" {
		t.Fatalf("Current() returned an empty field: %#v", got)
	}
}

func TestString(t *testing.T) {
	if got, want := String(), "version=dev commit=unknown buildDate=unknown"; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}
