package host

import (
	"sort"
	"strings"

	fuzzysearch "github.com/HobaiRiku/hosta/internal/search"
)

type MatchField string

const (
	MatchAlias       MatchField = "alias"
	MatchDisplayName MatchField = "display-name"
	MatchGroup       MatchField = "group"
	MatchHostName    MatchField = "host-name"
	MatchDescription MatchField = "description"
)

type Result struct {
	Host      Host
	Score     int
	Field     MatchField
	Value     string
	Positions []int
}

type fieldValue struct {
	field  MatchField
	value  string
	weight int
}

var fieldWeights = map[MatchField]int{
	MatchAlias:       100,
	MatchDisplayName: 90,
	MatchGroup:       60,
	MatchHostName:    50,
	MatchDescription: 30,
}

func (i *Index) Search(query string, limit int) []Result {
	query = strings.TrimSpace(query)
	if query == "" {
		return i.emptyResults(limit)
	}

	results := make([]Result, 0, len(i.hosts))
	for _, candidate := range i.hosts {
		best, matched := bestMatch(query, candidate)
		if matched {
			best.Host = clone(candidate)
			results = append(results, best)
		}
	}
	sort.SliceStable(results, func(left, right int) bool {
		if results[left].Score != results[right].Score {
			return results[left].Score > results[right].Score
		}
		return lessHost(results[left].Host, results[right].Host)
	})
	return limitResults(results, limit)
}

func (i *Index) emptyResults(limit int) []Result {
	results := make([]Result, len(i.hosts))
	for index, candidate := range i.hosts {
		results[index] = Result{Host: clone(candidate)}
	}
	return limitResults(results, limit)
}

func bestMatch(query string, candidate Host) (Result, bool) {
	fields := []fieldValue{
		{field: MatchAlias, value: candidate.Alias, weight: fieldWeights[MatchAlias]},
		{field: MatchDisplayName, value: candidate.DisplayName, weight: fieldWeights[MatchDisplayName]},
	}
	for _, alias := range candidate.Aliases {
		fields = append(fields, fieldValue{field: MatchAlias, value: alias, weight: fieldWeights[MatchAlias]})
	}
	fields = append(fields,
		fieldValue{field: MatchGroup, value: candidate.Group, weight: fieldWeights[MatchGroup]},
		fieldValue{field: MatchHostName, value: candidate.Preview.HostName, weight: fieldWeights[MatchHostName]},
		fieldValue{field: MatchDescription, value: candidate.Description, weight: fieldWeights[MatchDescription]},
	)

	var best Result
	matched := false
	for _, field := range fields {
		if field.value == "" {
			continue
		}
		match, ok := fuzzysearch.MatchOne(query, field.value)
		if !ok {
			continue
		}
		score := field.weight + match.Score + matchQualityBonus(query, field.value)
		if !matched || score > best.Score {
			best = Result{
				Score:     score,
				Field:     field.field,
				Value:     field.value,
				Positions: append([]int(nil), match.Positions...),
			}
			matched = true
		}
	}
	return best, matched
}

func matchQualityBonus(query, value string) int {
	query = strings.ToLower(query)
	value = strings.ToLower(value)
	switch {
	case value == query:
		return 1000
	case strings.HasPrefix(value, query):
		return 300
	case strings.Contains(value, query):
		return 100
	default:
		return 0
	}
}

func lessHost(left, right Host) bool {
	leftGroup := strings.ToLower(left.Group)
	rightGroup := strings.ToLower(right.Group)
	if leftGroup != rightGroup {
		return leftGroup < rightGroup
	}
	leftAlias := strings.ToLower(left.Alias)
	rightAlias := strings.ToLower(right.Alias)
	if leftAlias != rightAlias {
		return leftAlias < rightAlias
	}
	return left.Alias < right.Alias
}

func limitResults(results []Result, limit int) []Result {
	if limit > 0 && len(results) > limit {
		return results[:limit]
	}
	return results
}
