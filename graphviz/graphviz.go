package graphviz

import (
	"fmt"
	"github.com/awalterschulze/gographviz"
	"github.com/vgarvardt/x11colors-go"
	"golang.org/x/mod/semver"
	"sort"
	"strings"
	"terraform-vercheck/internals"
)

func versionUsed(version string, source *internals.Dependency,
	dependencies []*internals.Dependency) bool {

	for _, dep := range dependencies {
		if source.Name != dep.Name {
			continue
		}

		if semver.Compare(version, dep.CurrentVersion) == 0 ||
			semver.Compare(version, dep.LatestVersion) == 0 {
			return true
		}
	}
	return false
}

func modulesToDependencies(modules internals.Modules) []*internals.Dependency {
	deps := make([]*internals.Dependency, 0)

	for _, dep := range modules {
		deps = append(deps, &dep.Dependency)
	}

	return deps
}

func providersToDepdendencies(providers internals.Providers) []*internals.Dependency {
	deps := make([]*internals.Dependency, 0)

	for _, dep := range providers {
		deps = append(deps, &dep.Dependency)
	}

	return deps
}

func sanitizeVersion(version string) string {
	canonical := semver.Canonical(version)
	prerelease := semver.Prerelease(version)
	build := semver.Build(version)

	if prerelease != "" && build != "" {
		return fmt.Sprintf("%s-%s+%s", canonical, prerelease, build)
	} else if prerelease != "" && build == "" {
		return fmt.Sprintf("%s-%s", canonical, prerelease)
	} else if prerelease == "" && build != "" {
		return fmt.Sprintf("%s+%s", canonical, build)
	} else {
		return canonical
	}
}

func toPortID(version string) string {
	return fmt.Sprintf(" | <f%s> %s", version, version)
}

func createLabel(dep *internals.Dependency, deps []*internals.Dependency) string {
	out := fmt.Sprintf(`"<name> %s`, dep.Name)

	for _, version := range dep.Versions {
		sanitized := sanitizeVersion(version)
		if versionUsed(version, dep, deps) {
			out += toPortID(sanitized)
		}
	}

	return out + `"`
}

func getRandomColor(used map[*internals.Module]x11colors.X11Color) x11colors.X11Color {
	color := x11colors.Random()

	for _, usedColor := range used {
		if color == usedColor {
			return getRandomColor(used)
		}
	}

	return color
}

func associateChildren(graph *gographviz.Graph, srcNodeName, srcPortIdentifier string,
	children internals.ModuleToModuleAssociations, color string) {
	if len(children) == 0 {
		return
	}
	for _, child := range children {
		dstNodeName := fmt.Sprintf("\"%s\"", child.Module.Name)
		dstPortIdentifier := fmt.Sprintf("\"f%s\"", sanitizeVersion(child.Module.CurrentVersion))

		attrs := make(map[string]string)
		attrs["color"] = fmt.Sprintf("\"%s\"", color)
		graph.AddPortEdge(srcNodeName, srcPortIdentifier, dstNodeName, dstPortIdentifier, true, attrs)

		associateChildren(graph, dstNodeName, dstPortIdentifier, child.Children, color)
	}
}

func associateProviders(graph *gographviz.Graph, srcNodeName, srcPortIdentifier,
	color string, children []*internals.Provider) {

	for _, child := range children {
		dstNodeName := fmt.Sprintf("\"%s\"", child.Name)
		dstPortIdentifier := fmt.Sprintf("\"f%s\"", sanitizeVersion(child.CurrentVersion))

		attrs := make(map[string]string)
		attrs["color"] = fmt.Sprintf("\"%s\"", color)
		graph.AddPortEdge(srcNodeName, srcPortIdentifier, dstNodeName, dstPortIdentifier, true, attrs)
	}

}

func sortDependencyVerisons(deps []*internals.Dependency) {
	for i := range deps {
		sort.SliceStable(deps[i].Versions, func(x, y int) bool {
			return semver.Compare(deps[i].Versions[x], deps[i].Versions[y]) < 0
		})
	}
}

func moduleInBranch(module *internals.Module, assoc *internals.ModuleToModule) bool {
	if len(assoc.Children) == 0 {
		return false
	}

	if assoc.Module == module {
		return true
	}

	for _, child := range assoc.Children {
		if moduleInBranch(module, child) {
			return true
		}
	}

	return false
}

func getModuleColor(colors map[*internals.Module]x11colors.X11Color,
	associations internals.ModuleToModuleAssociations,
	module *internals.Module) x11colors.X11Color {

	for _, association := range associations {
		if association.Module != nil {
			continue
		}

		for _, child := range association.Children {
			color := colors[child.Module]

			if moduleInBranch(module, child) {
				return color
			}
		}
	}

	return x11colors.X11Color{}
}

// ToGraph : Create GraphViz DOT file representing the dependency graph.
func ToGraph(modules internals.Modules, providers internals.Providers,
	moduleToModuleAssociations internals.ModuleToModuleAssociations,
	moduleToProviderAssociations internals.ModuleToProviderAssocations) string {

	graph := gographviz.NewGraph()
	graph.SetName("G")
	graph.SetDir(true)
	graph.AddAttr("G", "rankdir", "LR")

	rootAttrs := make(map[string]string)
	rootAttrs["label"] = "\"<name> root | <flatest> latest\""
	rootAttrs["shape"] = "\"record\""
	graph.AddNode("G", "\"root\"", rootAttrs)

	dependencies := append(modulesToDependencies(modules),
		providersToDepdendencies(providers)...)

	sortDependencyVerisons(dependencies)

	for _, dep := range dependencies {
		label := createLabel(dep, dependencies)
		attrs := make(map[string]string)

		attrs["label"] = label
		attrs["shape"] = "\"record\""
		graph.AddNode("G", "\""+dep.Name+"\"", attrs)
	}

	colors := make(map[*internals.Module]x11colors.X11Color)

	for _, association := range moduleToModuleAssociations {
		if association.Module != nil {
			continue
		}

		srcNodeName := "\"root\""
		srcPortIdentifier := "\"flatest\""

		for _, child := range association.Children {
			color := getRandomColor(colors)
			colors[child.Module] = color
			colorName := strings.Replace(color.Name.Slugify(), "-", "", -1)

			dstNodeName := fmt.Sprintf("\"%s\"", child.Module.Name)
			dstPortIdentifier := fmt.Sprintf("\"f%s\"", child.Module.CurrentVersion)
			attrs := make(map[string]string)
			attrs["color"] = fmt.Sprintf("\"%s\"", colorName)
			graph.AddPortEdge(srcNodeName, srcPortIdentifier, dstNodeName, dstPortIdentifier, true, attrs)

			associateChildren(graph, dstNodeName, dstPortIdentifier,
				child.Children, colorName)
		}
	}

	for _, association := range moduleToProviderAssociations {
		var color x11colors.X11Color

		if association.Module == nil {
			color = x11colors.Black
		} else {
			color = getModuleColor(colors, moduleToModuleAssociations,
				association.Module)
		}

		colorName := strings.Replace(color.Name.Slugify(), "-", "", -1)

		var srcNodeName, srcPortIdentifier string

		if association.Module == nil {
			srcNodeName = `"root"`
			srcPortIdentifier = `"flatest"`
		} else {
			srcNodeName = fmt.Sprintf(`"%s"`, association.Module.Name)
			srcPortIdentifier = fmt.Sprintf(`"f%s"`, association.Module.CurrentVersion)
		}

		associateProviders(graph, srcNodeName, srcPortIdentifier, colorName,
			association.Children)
	}

	return graph.String()
}
