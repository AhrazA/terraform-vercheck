package main

import (
	"bufio"
	"flag"
	log "github.com/sirupsen/logrus"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"sync"
	"terraform-vercheck/extraction"
	"terraform-vercheck/graphviz"
	"terraform-vercheck/internals"
)

func getExitCode(modules []*internals.Module) int {
	exitCode := 0

	for _, module := range modules {
		if module.LatestVersion != module.CurrentVersion {
			exitCode = 1
		}
	}

	return exitCode
}

type discovery struct {
	parent   *internals.Module
	module   *internals.Module
	provider *internals.Provider
	depth    int
}

func parseRepository(directory, sshKeyFile string,
	fileRe, ignoreRe *regexp.Regexp, repoWg *sync.WaitGroup,
	discoveries chan<- discovery, parent *internals.Module, depth int) error {

	identifiers, err := extraction.ProcessDirectory(directory, fileRe, ignoreRe)

	if err != nil {
		return err
	}

	repoWg.Add(len(identifiers))

	for _, identifier := range identifiers {
		go func(id internals.Identifier) {
			defer repoWg.Done()
			module, provider, err := extraction.ExtractFromIdentifier(id, sshKeyFile)

			if err != nil {
				log.WithFields(log.Fields{
					"identifier": id,
					"error":      err,
				}).Warnf("Error extracting dependency")
			}

			discoveries <- discovery{
				parent:   parent,
				module:   module,
				provider: provider,
				depth:    depth,
			}
		}(identifier)
	}

	return nil
}

func pumpDiscoveries(buffer <-chan discovery, out chan<- discovery, maxDepth int,
	callback func(discovery)) {

	for new := range buffer {
		if new.depth <= maxDepth {
			callback(new)
		}
		out <- new
	}

	close(out)
}

func orchestrateRoutines(rootDirectory, sshKeyFilePath string,
	fileRe, ignoreRe *regexp.Regexp, discoveries chan<- discovery, maxDepth int) error {

	var repoWg sync.WaitGroup
	discoveryBuffer := make(chan discovery)

	err := parseRepository(rootDirectory, sshKeyFilePath, fileRe, ignoreRe,
		&repoWg, discoveryBuffer, nil, 0)

	if err != nil {
		return err
	}

	go pumpDiscoveries(discoveryBuffer, discoveries, maxDepth, func(new discovery) {
		if new.module != nil {
			log.Infof("Parsing submodule: %s", new.module.Name)

			err := parseRepository(new.module.Path, sshKeyFilePath, fileRe,
				ignoreRe, &repoWg, discoveryBuffer, new.module,
				new.depth+1)

			if err != nil {
				log.WithFields(log.Fields{
					"module": new.module,
					"error":  err,
				}).Error("Failed to parse repository")
			}
		}
	})

	go func() {
		repoWg.Wait()
		close(discoveryBuffer)
	}()

	return nil
}

func htmlTemplate(graph, htmlFilePath string) error {
	templ := template.Must(template.ParseFiles("templates/index.html"))

	f, err := os.Create(htmlFilePath)

	if err != nil {
		return err
	}

	writer := bufio.NewWriter(f)

	err = templ.Execute(writer, struct {
		DotGraph template.JS
	}{template.JS(graph)})

	if err != nil {
		return err
	}

	writer.Flush()

	return nil
}

func run(config config) int {
	log.Info("Running vercheck")

	if config.logFilePath != "" {
		logFile, err := os.OpenFile(config.logFilePath,
			os.O_CREATE|os.O_WRONLY, 0644)

		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatal("Error opening log file.")
		}

		defer logFile.Close()

		mw := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(mw)
	}

	if config.debug {
		log.SetLevel(log.DebugLevel)
	}

	fileRe := regexp.MustCompile(config.filePattern)
	ignoreRe := regexp.MustCompile(config.ignorePattern)

	modules := make(internals.Modules, 0)
	providers := make(internals.Providers, 0)

	moduleToModuleAssociations := make(internals.ModuleToModuleAssociations, 0)
	moduleToProviderAssociations := make(internals.ModuleToProviderAssocations, 0)

	discoveries := make(chan discovery)

	err := orchestrateRoutines(config.directory,
		config.sshKeyFilePath,
		fileRe,
		ignoreRe,
		discoveries,
		config.depth)

	if err != nil {
		log.Fatalf("Failed: %s", err)
	}

	for discovery := range discoveries {
		if discovery.module != nil {
			modules = modules.Add(discovery.module)
			moduleToModuleAssociations = moduleToModuleAssociations.Associate(
				discovery.parent, discovery.module)
		}

		if discovery.provider != nil {
			providers = providers.Add(discovery.provider)
			moduleToProviderAssociations = moduleToProviderAssociations.Associate(
				discovery.parent, discovery.provider)
		}
	}

	var dotGraph string

	if config.dotFilePath != "" || config.htmlFilePath != "" {
		dotGraph = graphviz.ToGraph(modules, providers,
			moduleToModuleAssociations, moduleToProviderAssociations)
	}

	if config.dotFilePath != "" {
		ioutil.WriteFile(config.dotFilePath, []byte(dotGraph), 0644)
	}

	if config.htmlFilePath != "" {
		err := htmlTemplate(dotGraph, config.htmlFilePath)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Error executing template")
		} else {
			log.WithFields(log.Fields{
				"path": config.htmlFilePath,
			}).Info("Created HTML.")
		}
	}

	return getExitCode(modules)
}

type config struct {
	directory      string
	debug          bool
	filePattern    string
	ignorePattern  string
	sshKeyFilePath string
	logFilePath    string
	dotFilePath    string
	htmlFilePath   string
	depth          int
}

func main() {
	directory := flag.String("directory", "./",
		"Specify the root terraform plan directory")
	debug := flag.Bool("debug", false,
		"Debug logging")
	filePattern := flag.String("pattern", `.+\.tf`,
		"Regex pattern to match target files")
	ignorePattern := flag.String("ignorepattern", `test`,
		"Regex pattern for directories to ignore")
	sshKeyFilePath := flag.String("key", "",
		"GitHub ssh key path")
	logFilePath := flag.String("log", "",
		"Output log file")
	dotFilePath := flag.String("graph", "",
		"Output graphviz DOT file path")
	htmlFilePath := flag.String("html", "",
		"Output HTML file path")
	depth := flag.Int("depth", 10,
		"Depth of submodules to evaluate")

	flag.Parse()

	config := config{
		directory:      *directory,
		debug:          *debug,
		filePattern:    *filePattern,
		ignorePattern:  *ignorePattern,
		sshKeyFilePath: *sshKeyFilePath,
		logFilePath:    *logFilePath,
		dotFilePath:    *dotFilePath,
		htmlFilePath:   *htmlFilePath,
		depth:          *depth,
	}

	os.Exit(run(config))
}
