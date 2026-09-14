package app

import (
	"github.com/HobaiRiku/hosta/internal/host"
	"github.com/HobaiRiku/hosta/internal/sshconfig"
)

type Snapshot struct {
	Config *sshconfig.Config
	Index  *host.Index
}

func Load(configPath string, options sshconfig.Options) (*Snapshot, error) {
	config, err := sshconfig.Parse(configPath, options)
	if err != nil {
		return nil, err
	}
	index, err := IndexFromSSHConfig(config)
	if err != nil {
		return nil, err
	}
	return &Snapshot{Config: config, Index: index}, nil
}

func (s *Snapshot) HasErrors() bool {
	if s == nil || s.Config == nil {
		return true
	}
	for _, diagnostic := range s.Config.Diagnostics {
		if diagnostic.Severity == sshconfig.SeverityError {
			return true
		}
	}
	return false
}
