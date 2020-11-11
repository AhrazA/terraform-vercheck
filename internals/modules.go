package internals

import (
	"fmt"
)

// Module : Terraform module dependency
type Module struct {
	Dependency
	Source string
	Path   string
}

func (m Module) String() string {
	return fmt.Sprintf("%s - %s", m.Source, m.CurrentVersion)
}

// Modules : Slice of Module
type Modules []*Module

// Add : Add a module
func (ms Modules) Add(module *Module) Modules {
	return append(ms, module)
}
