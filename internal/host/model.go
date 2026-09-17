package host

type Origin string

const (
	OriginNative  Origin = "native"
	OriginManaged Origin = "managed"
)

type Source struct {
	File string
	Line int
}

type Preview struct {
	HostName string
	User     string
	Port     string
}

type Host struct {
	ID          string
	Alias       string
	Aliases     []string
	DisplayName string
	Group       string
	Description string
	Preview     Preview
	Sources     []Source
	Origin      Origin
}

func clone(value Host) Host {
	value.Aliases = append([]string(nil), value.Aliases...)
	value.Sources = append([]Source(nil), value.Sources...)
	return value
}
