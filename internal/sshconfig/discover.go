package sshconfig

import (
	"sort"
	"strings"
)

func (c *Config) Hosts() []Host {
	hostsByAlias := make(map[string]*Host)
	blockHosts := make(map[int]*Host)
	hosts := make([]*Host, 0)

	for _, node := range c.Nodes {
		if node.Directive == "host" {
			aliases := make([]string, 0, len(node.Args))
			for _, pattern := range node.Args {
				if !isExplicitAlias(pattern) {
					continue
				}
				aliases = appendUniqueFold(aliases, pattern)
			}
			if len(aliases) == 0 {
				continue
			}

			var host *Host
			for _, alias := range aliases {
				if candidate := hostsByAlias[strings.ToLower(alias)]; candidate != nil {
					host = candidate
					break
				}
			}
			if host == nil {
				host = &Host{Alias: aliases[0]}
				hosts = append(hosts, host)
			}
			for _, alias := range aliases {
				if !strings.EqualFold(alias, host.Alias) {
					host.Aliases = appendUniqueFold(host.Aliases, alias)
				}
				hostsByAlias[strings.ToLower(alias)] = host
			}
			host.Sources = append(host.Sources, node.Source)
			blockHosts[node.Scope.ID] = host
			continue
		}
		if node.Scope.Kind != ScopeHost {
			continue
		}
		if host := blockHosts[node.Scope.ID]; host != nil {
			applyMetadata(host, node)
		}
	}

	result := make([]Host, len(hosts))
	for i, host := range hosts {
		result[i] = *host
		result[i].Aliases = append([]string(nil), host.Aliases...)
	}
	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i].Alias) < strings.ToLower(result[j].Alias)
	})
	return result
}

func applyMetadata(host *Host, node Node) {
	switch node.Directive {
	case "hosta.display-name", "displayname":
		if host.DisplayName == "" {
			host.DisplayName = strings.Join(node.Args, " ")
		}
	case "hosta.group", "group":
		if host.Group == "" {
			host.Group = strings.Join(node.Args, " ")
		}
	case "hosta.description", "description":
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
