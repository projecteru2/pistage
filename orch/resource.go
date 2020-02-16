package orch

// ResourceMetadata indicates resource metadata
type ResourceMetadata struct {
	CPU     float64
	CPUBind bool
	Memory  int64
	Image   string
	Podname string
	Count   int
	Network string
	Storage int64
	Volumes []string
	DNS     []string
	Env     []string
}
