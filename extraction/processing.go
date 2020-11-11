package extraction

import (
	"bufio"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"terraform-vercheck/internals"
)

type textProcessor interface {
	process(line string)
	extract() []internals.Identifier
}

func processLines(scanner *bufio.Scanner, processors []textProcessor) []internals.Identifier {
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		for _, proc := range processors {
			proc.process(line)
		}
	}

	dependencies := make([]internals.Identifier, 0)

	for _, proc := range processors {
		procDependencies := proc.extract()
		dependencies = append(dependencies, procDependencies...)
	}

	return dependencies
}

func extractIdentifiers(scanner *bufio.Scanner) []internals.Identifier {
	processors := make([]textProcessor, 2)
	processors[0] = &moduleIdentifierExtractor{
		inModule:   false,
		sourceURIs: make([]*string, 0),
	}
	processors[1] = &providerIdentifierExtractor{
		inRequiredProviders: false,
		providers:           make([]*providerIdentifier, 0),
	}

	identifiers := processLines(scanner, processors)
	return identifiers
}

// ProcessDirectory : Parse terraform files in a given directory and extract
//                    dependency identifiers
func ProcessDirectory(directory string, fileRe, ignoreRe *regexp.Regexp) ([]internals.Identifier, error) {
	identifiers := make([]internals.Identifier, 0)

	wd, err := os.Getwd()

	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"directory": wd,
	}).Debug("Working directory")

	err = filepath.Walk(directory,
		func(path string,
			info os.FileInfo,
			err error) error {

			if err != nil {
				return err
			}

			if info.Mode()&os.ModeSymlink == os.ModeSymlink {
				log.Debugf("Ignoring symlink: " + info.Name())
				return nil
			}

			if info.IsDir() {
				if info.Name() == ".git" || info.Name() == ".terraform" ||
					ignoreRe.MatchString(info.Name()) {

					log.Debug("Skipping directory: " + info.Name())
					return filepath.SkipDir
				}
			}

			if !fileRe.MatchString(info.Name()) {
				return nil
			}

			file, err := os.Open(path)

			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Fatal(err)
			}

			defer file.Close()

			scanner := bufio.NewScanner(file)
			identifiers = append(identifiers, extractIdentifiers(scanner)...)

			return nil
		})

	return identifiers, err
}
