package internals

import (
	"testing"
)

func TestAddModule(t *testing.T) {
	modules := make(Modules, 0)

	m1 := &Module{
		Dependency: Dependency{
			Name:           "Mod1",
			CurrentVersion: "v1",
		},
	}
	m2 := &Module{
		Dependency: Dependency{
			Name:           "Mod2",
			CurrentVersion: "v1",
		},
	}

	modules = modules.Add(m1)
	modules = modules.Add(m2)

	if len(modules) != 2 {
		t.Fatalf("Incorrect numbers of providers in Providers: %v", modules)
	}
}

func TestAddProviderAndIgnoreDuplicate(t *testing.T) {
	providers := make(Providers, 0)
	p1 := &Provider{
		Dependency: Dependency{
			Name:           "Test1",
			CurrentVersion: "v1",
		},
	}
	p2 := &Provider{
		Dependency: Dependency{
			Name:           "Test1",
			CurrentVersion: "v1",
		},
	}
	p3 := &Provider{
		Dependency: Dependency{
			Name:           "Test2",
			CurrentVersion: "v1",
			Versions: []string{
				"v2",
				"v3",
			},
		},
	}
	p4 := &Provider{
		Dependency: Dependency{
			Name:           "Test2",
			CurrentVersion: "v3",
		},
	}
	p5 := &Provider{
		Dependency: Dependency{
			Name:           "Test2",
			CurrentVersion: "v4",
		},
	}
	p6 := &Provider{
		Dependency: Dependency{
			Name:           "Test3",
			CurrentVersion: "v4",
		},
	}

	providers = providers.Add(p1)
	providers = providers.Add(p2)
	providers = providers.Add(p3)
	providers = providers.Add(p4)
	providers = providers.Add(p5)
	providers = providers.Add(p6)

	if len(providers) != 6 {
		t.Fatalf("Incorrect numbers of providers in Providers: %v", providers)
	}
}

func mtmAssociationExists(mtmas ModuleToModuleAssociations, parent *Module,
	child *Module) bool {

	childContains := func(children ModuleToModuleAssociations, target *Module) bool {
		for _, child := range children {
			if child.Module == target {
				return true
			}
		}
		return false
	}

	for _, mtma := range mtmas {
		if mtma.Module == parent {
			if childContains(mtma.Children, child) {
				return true
			}
		}
	}

	return false
}

func TestModuleToModuleAssociations(t *testing.T) {
	mtmas := make(ModuleToModuleAssociations, 0)

	m1 := &Module{
		Dependency: Dependency{
			Name:           "Mod1",
			CurrentVersion: "v1",
		},
	}
	m2 := &Module{
		Dependency: Dependency{
			Name:           "Mod2",
			CurrentVersion: "v1",
		},
	}
	m3 := &Module{
		Dependency: Dependency{
			Name:           "Mod3",
			CurrentVersion: "v3",
		},
	}

	assocs := []struct {
		parent *Module
		child  *Module
	}{
		{nil, m1},
		{m1, m2},
		{m2, m3},
		{m1, m3},
	}

	for _, assoc := range assocs {
		mtmas = mtmas.Associate(assoc.parent, assoc.child)
	}

	for _, assoc := range assocs {
		if !mtmAssociationExists(mtmas, assoc.parent, assoc.child) {
			t.Fatalf("%v not associated to %v", assoc.parent, assoc.child)
		}
	}
}

func mtpAssociationExists(mtmas ModuleToProviderAssocations, parent *Module,
	provider *Provider) bool {

	childContains := func(children Providers, target *Provider) bool {
		for _, child := range children {
			if child == target {
				return true
			}
		}
		return false
	}

	for _, mtma := range mtmas {
		if mtma.Module == parent {
			if childContains(mtma.Children, provider) {
				return true
			}
		}
	}

	return false
}

func TestModuleToProviderAssociations(t *testing.T) {
	mtpas := make(ModuleToProviderAssocations, 0)

	m1 := &Module{
		Dependency: Dependency{
			Name:           "Mod1",
			CurrentVersion: "v1",
		},
	}
	m2 := &Module{
		Dependency: Dependency{
			Name:           "Mod2",
			CurrentVersion: "v1",
		},
	}
	p1 := &Provider{
		Dependency: Dependency{
			Name:           "Test1",
			CurrentVersion: "v1",
		},
	}
	p2 := &Provider{
		Dependency: Dependency{
			Name:           "Test1",
			CurrentVersion: "v1",
		},
	}

	assocs := []struct {
		parent   *Module
		provider *Provider
	}{
		{m1, p1},
		{m1, p2},
		{m2, p2},
	}

	for _, assoc := range assocs {
		mtpas = mtpas.Associate(assoc.parent, assoc.provider)
	}

	for _, assoc := range assocs {
		if !mtpAssociationExists(mtpas, assoc.parent, assoc.provider) {
			t.Fatalf("%v not associated to %v", assoc.parent, assoc.provider)
		}
	}
}
