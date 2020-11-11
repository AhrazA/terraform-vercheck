package graphviz

import (
	"github.com/andreyvit/diff"
	"terraform-vercheck/internals"
	"testing"
)

func TestToGraph(t *testing.T) {
	expected := `
digraph G {
        rankdir=LR;
        "Mod1":"fv1"->"Dep1":"fv1.0.0"[ color="" ];
        "Mod1":"fv1"->"Dep1":"fv1.0.0"[ color="" ];
        "Mod1":"fv1"->"Dep2":"fv1.0.0"[ color="" ];
        "Mod2":"fv1"->"Dep2":"fv1.0.0"[ color="" ];
        "Dep1" [ label="<name> Dep1", shape="record" ];
        "Dep2" [ label="<name> Dep2", shape="record" ];
        "Mod1" [ label="<name> Mod1 | <fv1.0.0> v1.0.0 | <fv3.0.0> v3.0.0", shape="record" ];
        "Mod2" [ label="<name> Mod2 | <fv1.0.0> v1.0.0 | <fv4.0.0> v4.0.0", shape="record" ];
        "root" [ label="<name> root | <flatest> latest", shape="record" ];

}`

	modules := make(internals.Modules, 0)
	m1 := &internals.Module{
		Dependency: internals.Dependency{
			Name:           "Mod1",
			CurrentVersion: "v1",
			Versions:       []string{"v1", "v2", "v3"},
			LatestVersion:  "v3",
		},
	}
	m2 := &internals.Module{
		Dependency: internals.Dependency{
			Name:           "Mod2",
			CurrentVersion: "v1",
			Versions:       []string{"v1", "v2", "v3", "v4"},
			LatestVersion:  "v4",
		},
	}

	modules = modules.Add(m1)
	modules = modules.Add(m2)

	providers := make(internals.Providers, 0)
	p1 := &internals.Provider{
		Dependency: internals.Dependency{
			Name:           "Dep1",
			CurrentVersion: "v1",
		},
	}
	p2 := &internals.Provider{
		Dependency: internals.Dependency{
			Name:           "Dep1",
			CurrentVersion: "v1",
		},
	}
	p3 := &internals.Provider{
		Dependency: internals.Dependency{
			Name:           "Dep2",
			CurrentVersion: "v1",
			Versions: []string{
				"v2",
				"v3",
			},
		},
	}

	providers = providers.Add(p1)
	providers = providers.Add(p2)
	providers = providers.Add(p3)

	mtmAssociations := make(internals.ModuleToModuleAssociations, 0)
	mtmAssociations = mtmAssociations.Associate(m1, m2)

	mtpAssociations := make(internals.ModuleToProviderAssocations, 0)
	mtpAssociations = mtpAssociations.Associate(m1, p1)
	mtpAssociations = mtpAssociations.Associate(m1, p2)
	mtpAssociations = mtpAssociations.Associate(m1, p3)
	mtpAssociations = mtpAssociations.Associate(m2, p3)

	graph := ToGraph(modules, providers, mtmAssociations, mtpAssociations)

	if diff.TrimLinesInString(graph) != diff.TrimLinesInString(expected) {
		t.Errorf("Invalid graph generated: %v", diff.LineDiff(expected, graph))
	}
}
