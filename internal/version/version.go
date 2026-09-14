package version

import "fmt"

// Variables are injected at build time via -ldflags.
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"buildDate"`
}

func Current() Info {
	return Info{Version: Version, Commit: Commit, BuildDate: BuildDate}
}

func String() string {
	return fmt.Sprintf("version=%s commit=%s buildDate=%s", Version, Commit, BuildDate)
}
