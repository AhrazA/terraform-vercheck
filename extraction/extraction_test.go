package extraction

import (
	"bufio"
	"bytes"
	"terraform-vercheck/internals"
	"testing"
)

func TestExtractIdentifiers(t *testing.T) {
	buf := bytes.NewBufferString(`module "silo_base" {
	source = "git::ssh://git@github.com/AhrazA/Infrastructure/somerepo.git?ref=v3.2.0"
}

module "silo_base" {
  source = "git::ssh://git@github.com/AhrazA/Infrastructure/somerepo.git?ref=v3.1.0"
}

module "silo_base" {
	source = "git::ssh://git@github.com/AhrazA/Infrastructure/somerepo.git?ref=v3.0.0"
}

terraform {
  required_version = "~> 0.12"
  required_providers {
    azurerm = "~> 1.41"
    azuread = "~> 0.6"
    helm = "~> 0.10"
  }
}`)

	expectedProviders := []providerIdentifier{
		{
			provider: "azurerm",
			version:  "v1.41",
		},
		{
			provider: "azuread",
			version:  "v0.6",
		},
		{
			provider: "helm",
			version:  "v0.10",
		},
	}

	expectedModules := []moduleIdentifier{
		{
			sourceURI: "git::ssh://git@github.com/AhrazA/Infrastructure/somerepo.git?ref=v3.2.0",
		},
		{
			sourceURI: "git::ssh://git@github.com/AhrazA/Infrastructure/somerepo.git?ref=v3.1.0",
		},
		{
			sourceURI: "git::ssh://git@github.com/AhrazA/Infrastructure/somerepo.git?ref=v3.0.0",
		},
	}

	scanner := bufio.NewScanner(buf)
	identifiers := extractIdentifiers(scanner)

	expectedProvidersContain := func(pid *providerIdentifier) bool {
		for _, ep := range expectedProviders {
			if ep.provider == pid.provider &&
				ep.version == pid.version {
				return true
			}
		}
		return false
	}

	expectedModulesContain := func(mid *moduleIdentifier) bool {
		for _, mi := range expectedModules {
			if mi.sourceURI == mid.sourceURI {
				return true
			}
		}
		return false
	}

	for _, id := range identifiers {
		if id.GetDependencyType() == internals.ProviderDependency {
			pid, ok := id.(*providerIdentifier)

			if !ok {
				t.Error("Invalid dependency type for identifier")
			}

			if !expectedProvidersContain(pid) {
				t.Errorf("Identifier %v is not correct.", pid)
			}
		} else if id.GetDependencyType() == internals.ModuleDependency {
			mid, ok := id.(*moduleIdentifier)

			if !ok {
				t.Error("Invalid dependency type for identifier")
			}

			if !expectedModulesContain(mid) {
				t.Errorf("Identifier %v is not correct.", mid)
			}
		}
	}
}

type dummyTextProcessor struct {
	processed int
	extracted int
	textProcessor
}

func (dtp *dummyTextProcessor) process(line string) {
	dtp.processed++
}

func (dtp *dummyTextProcessor) extract() []internals.Identifier {
	dtp.extracted++
	return nil
}

func TestProcessLines(t *testing.T) {
	buf := bytes.NewBufferString(`l1
	l2
	l3
	l4
	l5`)

	dtp := dummyTextProcessor{}
	scanner := bufio.NewScanner(buf)

	processLines(scanner, []textProcessor{&dtp})

	if dtp.processed != 5 {
		t.Errorf("Did not process correct amount of lines: %d", dtp.processed)
	}

	if dtp.extracted != 1 {
		t.Errorf("Did not extract correct amount of lines: %d", dtp.extracted)
	}
}

// TODO
func TestExtractFromIdentifier(t *testing.T) {

}
