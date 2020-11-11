package extraction

import (
	"fmt"
	"terraform-vercheck/git"
	"terraform-vercheck/internals"
)

// ExtractFromIdentifier : Extract a module or provider from its identifier
func ExtractFromIdentifier(identifier internals.Identifier,
	sshKeyFile string) (*internals.Module, *internals.Provider, error) {

	switch identifierType := identifier.GetDependencyType(); identifierType {

	case internals.ModuleDependency:
		moduleIdentifier, _ := identifier.(*moduleIdentifier)
		moduleDependency, err := extractModule(*moduleIdentifier, sshKeyFile)
		return moduleDependency, nil, err

	case internals.ProviderDependency:
		providerIdentifier, _ := identifier.(*providerIdentifier)
		providerDependency, err := extractProvider(*providerIdentifier)
		return nil, providerDependency, err

	default:
		return nil, nil, fmt.Errorf("unknown dependency type encountered: %d",
			identifierType)
	}
}

func extractModule(identifier moduleIdentifier,
	sshKeyFile string) (*internals.Module, error) {

	module, err := git.EvaluateGitModule(identifier.sourceURI,
		sshKeyFile)

	if err != nil {
		return nil, err
	}

	return module, nil
}

func extractProvider(identifier providerIdentifier) (*internals.Provider,
	error) {
	versions, latestVersion, err := getProviderVersions(identifier.provider)

	if err != nil {
		return nil, err
	}

	provider := internals.Provider{
		Dependency: internals.Dependency{
			CurrentVersion: identifier.version,
			Name:           identifier.provider,
			Versions:       versions,
			LatestVersion:  latestVersion,
		},
	}

	return &provider, nil
}
