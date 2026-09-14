package host

import (
	"reflect"
	"testing"
)

func TestNewIndexSortsByGroupAndAlias(t *testing.T) {
	index, err := NewIndex([]Host{
		{Alias: "zulu", Group: "work"},
		{Alias: "beta", Group: "personal"},
		{Alias: "alpha", Group: "personal"},
	})
	if err != nil {
		t.Fatalf("NewIndex() error = %v", err)
	}
	got := aliases(index.All())
	if want := []string{"alpha", "beta", "zulu"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("aliases = %#v, want %#v", got, want)
	}
}

func TestNewIndexRejectsDuplicateAlias(t *testing.T) {
	_, err := NewIndex([]Host{{Alias: "Home"}, {Alias: "home"}})
	if err == nil {
		t.Fatal("NewIndex() error = nil, want duplicate error")
	}
}

func TestIndexOwnsCopies(t *testing.T) {
	input := []Host{{Alias: "home", Tags: []string{"personal"}}}
	index, err := NewIndex(input)
	if err != nil {
		t.Fatalf("NewIndex() error = %v", err)
	}
	input[0].Tags[0] = "changed"
	result := index.All()
	result[0].Tags[0] = "also-changed"
	if got := index.All()[0].Tags[0]; got != "personal" {
		t.Fatalf("stored tag = %q, want personal", got)
	}
}

func aliases(hosts []Host) []string {
	result := make([]string, len(hosts))
	for i := range hosts {
		result[i] = hosts[i].Alias
	}
	return result
}
