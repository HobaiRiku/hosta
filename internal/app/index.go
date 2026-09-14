package app

import (
	"fmt"
	"strings"

	"github.com/HobaiRiku/hosta/internal/host"
	"github.com/HobaiRiku/hosta/internal/sshconfig"
)

func IndexFromSSHConfig(config *sshconfig.Config) (*host.Index, error) {
	if config == nil {
		return nil, fmt.Errorf("SSH config is nil")
	}
	discovered := config.Hosts()
	hosts := make([]host.Host, len(discovered))
	for i, value := range discovered {
		sources := make([]host.Source, len(value.Sources))
		for sourceIndex, source := range value.Sources {
			sources[sourceIndex] = host.Source{File: source.File, Line: source.Line}
		}
		hosts[i] = host.Host{
			ID:          "native:" + strings.ToLower(value.Alias),
			Alias:       value.Alias,
			Aliases:     append([]string(nil), value.Aliases...),
			DisplayName: value.DisplayName,
			Group:       value.Group,
			Tags:        append([]string(nil), value.Tags...),
			Description: value.Description,
			Preview: host.Preview{
				HostName: value.HostName,
				User:     value.User,
				Port:     value.Port,
			},
			Sources: sources,
			Origin:  host.OriginNative,
		}
	}
	return host.NewIndex(hosts)
}
