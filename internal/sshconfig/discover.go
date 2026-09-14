package sshconfig

import (
	"sort"
	"strings"
)

func (c *Config) Hosts() []Host {
	hosts := make(map[string]*Host)
	blockAliases := make(map[int][]string)

	for _, node := range c.Nodes {
		if node.Directive == "host" {
			for _, pattern := range node.Args {
				if !isExplicitAlias(pattern) {
					continue
				}
				alias := pattern
				key := strings.ToLower(alias)
				host, exists := hosts[key]
				if !exists {
					host = &Host{Alias: alias}
					hosts[key] = host
				}
				host.Sources = append(host.Sources, node.Source)
				blockAliases[node.Scope.ID] = appendUniqueFold(blockAliases[node.Scope.ID], alias)
			}
			continue
		}
		if node.Scope.Kind != ScopeHost {
			continue
		}
		for _, alias := range blockAliases[node.Scope.ID] {
			host := hosts[strings.ToLower(alias)]
			applyMetadata(host, node)
		}
	}

	result := make([]Host, 0, len(hosts))
	for _, host := range hosts {
		result = append(result, *host)
	}
	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i].Alias) < strings.ToLower(result[j].Alias)
	})
	return result
}

func applyMetadata(host *Host, node Node) {
	switch node.Directive {
	case "hosta.display-name":
		if host.DisplayName == "" {
			host.DisplayName = strings.Join(node.Args, " ")
		}
	case "hosta.group":
		if host.Group == "" {
			host.Group = strings.Join(node.Args, " ")
		}
	case "hosta.tags":
		for _, tag := range node.Args {
			host.Tags = appendUniqueFold(host.Tags, tag)
		}
	case "hosta.description":
		if host.Description == "" {
			host.Description = strings.Join(node.Args, " ")
		}
	case "hostname":
		if host.HostName == "" {
			host.HostName = strings.Join(node.Args, " ")
		}
	case "user":
		if host.User == "" {
			host.User = strings.Join(node.Args, " ")
		}
	case "port":
		if host.Port == "" {
			host.Port = strings.Join(node.Args, " ")
		}
	}
}

func isExplicitAlias(pattern string) bool {
	return pattern != "" && !strings.HasPrefix(pattern, "!") && !strings.ContainsAny(pattern, "*?[]")
}

func appendUniqueFold(values []string, value string) []string {
	for _, existing := range values {
		if strings.EqualFold(existing, value) {
			return values
		}
	}
	return append(values, value)
}
