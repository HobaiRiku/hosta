package host

import (
	"fmt"
	"sort"
	"strings"
)

type Index struct {
	hosts []Host
}

func NewIndex(hosts []Host) (*Index, error) {
	seen := make(map[string]string, len(hosts))
	indexed := make([]Host, len(hosts))
	for i, value := range hosts {
		value.Alias = strings.TrimSpace(value.Alias)
		if value.Alias == "" {
			return nil, fmt.Errorf("host at index %d has an empty alias", i)
		}
		value.Aliases = normalizeAliases(value.Alias, value.Aliases)
		for _, alias := range append([]string{value.Alias}, value.Aliases...) {
			key := strings.ToLower(alias)
			if existing, ok := seen[key]; ok {
				return nil, fmt.Errorf("duplicate host alias %q conflicts with %q", alias, existing)
			}
			seen[key] = value.Alias
		}
		indexed[i] = clone(value)
	}
	sortHosts(indexed)
	return &Index{hosts: indexed}, nil
}

func (i *Index) Len() int {
	return len(i.hosts)
}

func (i *Index) All() []Host {
	result := make([]Host, len(i.hosts))
	for index, value := range i.hosts {
		result[index] = clone(value)
	}
	return result
}

func (i *Index) Get(alias string) (Host, bool) {
	for _, candidate := range i.hosts {
		if strings.EqualFold(candidate.Alias, alias) {
			return clone(candidate), true
		}
		for _, alternate := range candidate.Aliases {
			if strings.EqualFold(alternate, alias) {
				return clone(candidate), true
			}
		}
	}
	return Host{}, false
}

func normalizeAliases(primary string, aliases []string) []string {
	result := make([]string, 0, len(aliases))
	for _, alias := range aliases {
		alias = strings.TrimSpace(alias)
		if alias == "" || strings.EqualFold(alias, primary) {
			continue
		}
		duplicate := false
		for _, existing := range result {
			if strings.EqualFold(existing, alias) {
				duplicate = true
				break
			}
		}
		if !duplicate {
			result = append(result, alias)
		}
	}
	return result
}

func sortHosts(hosts []Host) {
	sort.SliceStable(hosts, func(i, j int) bool {
		leftGroup := strings.ToLower(hosts[i].Group)
		rightGroup := strings.ToLower(hosts[j].Group)
		if leftGroup != rightGroup {
			return leftGroup < rightGroup
		}
		leftAlias := strings.ToLower(hosts[i].Alias)
		rightAlias := strings.ToLower(hosts[j].Alias)
		if leftAlias != rightAlias {
			return leftAlias < rightAlias
		}
		return hosts[i].Alias < hosts[j].Alias
	})
}
