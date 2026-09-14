package host

import (
	"reflect"
	"testing"
)

func TestSearchUsesAllFields(t *testing.T) {
	index := mustIndex(t, []Host{
		{Alias: "mt2500", DisplayName: "MT2500 路由器", Group: "personal", Tags: []string{"home", "router"}},
		{Alias: "prod-api", Description: "Primary production service", Preview: Preview{HostName: "api.example.com"}},
	})

	tests := []struct {
		query string
		alias string
		field MatchField
	}{
		{query: "路由", alias: "mt2500", field: MatchDisplayName},
		{query: "router", alias: "mt2500", field: MatchTag},
		{query: "example", alias: "prod-api", field: MatchHostName},
		{query: "production", alias: "prod-api", field: MatchDescription},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			results := index.Search(tt.query, 0)
			if len(results) == 0 || results[0].Host.Alias != tt.alias || results[0].Field != tt.field {
				t.Fatalf("Search(%q) = %#v", tt.query, results)
			}
		})
	}
}

func TestSearchPrefersExactThenFieldWeight(t *testing.T) {
	index := mustIndex(t, []Host{
		{Alias: "prod-long-alias", DisplayName: "Production"},
		{Alias: "production", DisplayName: "Other"},
		{Alias: "misc", Description: "production"},
	})
	results := index.Search("production", 0)
	if got, want := resultAliases(results), []string{"production", "prod-long-alias", "misc"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("result aliases = %#v, want %#v", got, want)
	}
}

func TestSearchIsDeterministicAndLimited(t *testing.T) {
	index := mustIndex(t, []Host{{Alias: "beta"}, {Alias: "alpha"}, {Alias: "alpine"}})
	first := resultAliases(index.Search("a", 2))
	second := resultAliases(index.Search("a", 2))
	if !reflect.DeepEqual(first, second) || len(first) != 2 {
		t.Fatalf("searches = %#v and %#v, want same two results", first, second)
	}
}

func TestEmptySearchUsesIndexOrder(t *testing.T) {
	index := mustIndex(t, []Host{{Alias: "work", Group: "b"}, {Alias: "home", Group: "a"}})
	if got, want := resultAliases(index.Search("", 0)), []string{"home", "work"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("result aliases = %#v, want %#v", got, want)
	}
}

func BenchmarkSearch(b *testing.B) {
	hosts := make([]Host, 500)
	for i := range hosts {
		hosts[i] = Host{
			Alias:       "production-host-" + string(rune('a'+i%26)) + string(rune('a'+i/26%26)),
			DisplayName: "Production Service",
			Group:       "work",
			Tags:        []string{"linux", "production"},
			Description: "Application server",
		}
	}
	index, err := NewIndex(hosts)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for range b.N {
		_ = index.Search("prod", 20)
	}
}

func mustIndex(t *testing.T, hosts []Host) *Index {
	t.Helper()
	index, err := NewIndex(hosts)
	if err != nil {
		t.Fatalf("NewIndex() error = %v", err)
	}
	return index
}

func resultAliases(results []Result) []string {
	aliases := make([]string, len(results))
	for i := range results {
		aliases[i] = results[i].Host.Alias
	}
	return aliases
}
