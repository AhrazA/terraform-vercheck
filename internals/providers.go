package internals

import (
	"fmt"
)

// Provider : A terraform provider (azurerm, helm, etc) dependency
type Provider struct {
	Dependency
}

func (p Provider) String() string {
	return fmt.Sprintf("%s - %s", p.Name, p.CurrentVersion)
}

// Providers : Slice of Provider
type Providers []*Provider

// Add : Add a provider
func (ps Providers) Add(provider *Provider) Providers {
	return append(ps, provider)
}
