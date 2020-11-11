package internals

const (
	// ModuleDependency identifier
	ModuleDependency = iota
	// ProviderDependency identifier
	ProviderDependency = iota
)

// Dependency : A semantically versioned terraform module dependency
type Dependency struct {
	CurrentVersion string
	LatestVersion  string
	Versions       []string
	Name           string
}

// Identifier : Identify what type of dependency
type Identifier interface {
	GetDependencyType() int
}
