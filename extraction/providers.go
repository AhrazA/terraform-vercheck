package extraction

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/mod/semver"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"terraform-vercheck/internals"
)

type providerIdentifierExtractor struct {
	inRequiredProviders bool
	providers           []*providerIdentifier
}

type providerIdentifier struct {
	provider string
	version  string
}

func (pi providerIdentifier) String() string {
	return fmt.Sprintf("ProviderIdentifier: %s - %s", pi.provider, pi.version)
}

func (pi *providerIdentifier) GetDependencyType() int {
	return internals.ProviderDependency
}

func (pie *providerIdentifierExtractor) process(line string) {
	const requiredProvidersPattern = "required_providers"
	const providerVersionPattern = `(.+) ?=.+ (\d*\.*\d*\.*\d*)\"`

	requiredProvidersRe := regexp.MustCompile(requiredProvidersPattern)
	providerVersionRe := regexp.MustCompile(providerVersionPattern)

	if !pie.inRequiredProviders {
		pie.inRequiredProviders = requiredProvidersRe.MatchString(line)
	}

	if pie.inRequiredProviders && providerVersionRe.MatchString(line) {
		providerVersions := providerVersionRe.FindStringSubmatch(line)

		if len(providerVersions) != 3 {
			log.Warnf("Failed to parse provider version specification correctly, line: %s", line)
		}

		version := strings.TrimSpace(providerVersions[2])

		// Prepend a "v" to adhere to Semver
		if version[0] != 'v' {
			version = fmt.Sprintf("v%s", version)
		}

		providerID := providerIdentifier{
			provider: strings.TrimSpace(providerVersions[1]),
			version:  version,
		}

		pie.providers = append(pie.providers, &providerID)
	}

	if pie.inRequiredProviders && line == "}" {
		pie.inRequiredProviders = false
	}
}

func (pie *providerIdentifierExtractor) extract() []internals.Identifier {
	ret := make([]internals.Identifier, len(pie.providers))

	for i, provider := range pie.providers {
		ret[i] = provider
	}

	return ret
}

type terraformRegistryVersion struct {
	Version   string
	Protocols []string
	Platforms []struct {
		Os   string
		Arch string
	}
}

type terraformRegistryVersionResp struct {
	ID       string
	Versions []terraformRegistryVersion
	Warnings interface{}
}

// https://www.terraform.io/docs/internals/provider-registry-protocol.html
func getProviderVersions(provider string) ([]string, string, error) {
	providerRegistryURI := fmt.Sprintf("https://registry.terraform.io/v1/providers/hashicorp/%s/versions", provider)

	resp, err := http.Get(providerRegistryURI)
	ret := make([]string, 0)

	if err != nil {
		return nil, "", err
	}

	respBody, err := ioutil.ReadAll(resp.Body)

	var respData terraformRegistryVersionResp
	json.Unmarshal(respBody, &respData)

	if err != nil {
		return nil, "", err
	}

	latest := "v0.0.0"
	for _, version := range respData.Versions {
		// The SemVer library expects versions to be prepended with "v"
		semverVersion := "v" + version.Version
		switch semver.Compare(semverVersion, latest) {
		case 1:
			latest = semverVersion
		default:
		}
		ret = append(ret, semverVersion)
	}

	log.WithFields(log.Fields{
		"version":  latest,
		"provider": provider,
		"versions": ret,
	}).Debug("Found latest version")

	return ret, latest, nil
}
