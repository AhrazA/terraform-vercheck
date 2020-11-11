package internals

// ModuleToModule : Terraform submodule dependency
//                  A nil value for the parent indicates the root level module.
type ModuleToModule struct {
	Module   *Module
	Children ModuleToModuleAssociations
}

// ModuleToModuleAssociations : Slice of ModuleToModule, representing the entire
//															graph.
type ModuleToModuleAssociations []*ModuleToModule

// Associate : Associate a module to a module. Creates a new ModuleToModule if
//             one does not already exist.
func (mtms ModuleToModuleAssociations) Associate(parent, child *Module) ModuleToModuleAssociations {
	childMtm := ModuleToModule{
		Module:   child,
		Children: make(ModuleToModuleAssociations, 0),
	}

	for i, mtm := range mtms {
		if mtm.Module == parent {
			mtms[i].Children = append(mtms[i].Children, &childMtm)
			return append(mtms, &childMtm)
		}
	}

	mtm := ModuleToModule{
		Module:   parent,
		Children: make(ModuleToModuleAssociations, 0),
	}

	mtm.Children = append(mtm.Children, &childMtm)
	return append(mtms, &mtm, &childMtm)
}

// ModuleToProvider : Terraform module provider dependency
type ModuleToProvider struct {
	Module   *Module
	Children Providers
}

// ModuleToProviderAssocations : Slice of ModuleToProvider, representing the
//															 entire graph.
type ModuleToProviderAssocations []ModuleToProvider

// Associate : Associate a provider to a module. Creates a new ModuleToProvider
//             if one does not already exist.
func (mtps ModuleToProviderAssocations) Associate(parent *Module,
	provider *Provider) ModuleToProviderAssocations {

	for i, mtp := range mtps {
		if mtp.Module == parent {
			mtps[i].Children = append(mtps[i].Children, provider)
			return mtps
		}
	}

	mtp := ModuleToProvider{
		Module:   parent,
		Children: make(Providers, 0),
	}

	mtp.Children = append(mtp.Children, provider)
	return append(mtps, mtp)
}
