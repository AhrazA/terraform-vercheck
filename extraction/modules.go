package extraction

import (
	log "github.com/sirupsen/logrus"
	"regexp"
	"terraform-vercheck/internals"
)

type moduleIdentifierExtractor struct {
	inModule   bool
	sourceURIs []*string
}

type moduleIdentifier struct {
	sourceURI string
}

func (mi moduleIdentifier) String() string {
	return "ModuleIdentifier: " + mi.sourceURI
}

func (mi *moduleIdentifier) GetDependencyType() int {
	return internals.ModuleDependency
}

func (mdp *moduleIdentifierExtractor) process(line string) {
	const moduleIdentifierPattern = `^module.+{$`
	const sourceAttributeIdentifierPattern = `^source`
	const stringContentPattern = `"[^"]*"`

	moduleRe := regexp.MustCompile(moduleIdentifierPattern)
	sourceRe := regexp.MustCompile(sourceAttributeIdentifierPattern)
	stringRe := regexp.MustCompile(stringContentPattern)

	if !mdp.inModule {
		mdp.inModule = moduleRe.MatchString(line)
	}

	if mdp.inModule && sourceRe.MatchString(line) {
		sourceURI := stringRe.FindAllString(line, 1)[0]
		sourceURI = sourceURI[1 : len(sourceURI)-1]
		mdp.sourceURIs = append(mdp.sourceURIs, &sourceURI)
	}

	if mdp.inModule && line == "}" {
		mdp.inModule = false
	}
}

func (mdp *moduleIdentifierExtractor) extract() []internals.Identifier {
	const directoryModuleSourcePattern = `^(\./)|^(\.\./)`
	directorySourceRe := regexp.MustCompile(directoryModuleSourcePattern)

	identifiers := make([]internals.Identifier, 0)

	for _, uri := range mdp.sourceURIs {
		if *uri == "" {
			log.Debug("No source URI parsed yet. Probably none present.")
			continue
		}
		if directorySourceRe.MatchString(*uri) {
			log.Debug("File based module source paths not supported: ", *uri)
			continue
		}

		moduleIdentifier := moduleIdentifier{
			sourceURI: *uri,
		}

		identifiers = append(identifiers, &moduleIdentifier)
	}

	return identifiers
}
