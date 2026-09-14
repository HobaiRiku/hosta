package search

import "github.com/sahilm/fuzzy"

type Match struct {
	Index     int
	Score     int
	Positions []int
}

func MatchOne(query, candidate string) (Match, bool) {
	matches := fuzzy.Find(query, []string{candidate})
	if len(matches) == 0 {
		return Match{}, false
	}
	return Match{
		Index:     0,
		Score:     matches[0].Score,
		Positions: append([]int(nil), matches[0].MatchedIndexes...),
	}, true
}

func Find(query string, candidates []string) []Match {
	if query == "" {
		matches := make([]Match, len(candidates))
		for i := range candidates {
			matches[i] = Match{Index: i}
		}
		return matches
	}

	found := fuzzy.Find(query, candidates)
	matches := make([]Match, len(found))
	for i, item := range found {
		matches[i] = Match{
			Index:     item.Index,
			Score:     item.Score,
			Positions: append([]int(nil), item.MatchedIndexes...),
		}
	}
	return matches
}
