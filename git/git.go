package git

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/rs/xid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/mod/semver"
	"regexp"
	"strings"
	"terraform-vercheck/internals"
)

func cloneGitRepo(uri string, sshKey *ssh.PublicKeys) (*git.Repository, string, error) {
	log.Debugf("Cloning: %s", uri)
	guid := xid.New().String()
	clonePath := "/tmp/tfvercheck/" + guid

	repo, err := git.PlainClone(clonePath, false, &git.CloneOptions{
		URL:  uri,
		Auth: sshKey,
	})

	return repo, clonePath, err
}

func getVersions(repo *git.Repository) ([]string, string, error) {
	tags, err := repo.Tags()

	if err != nil {
		return nil, "", err
	}

	latestVersion := "v0.0.0"
	versions := make([]string, 0)

	err = tags.ForEach(func(t *plumbing.Reference) error {
		version := strings.Replace(t.Name().String(), "refs/tags/", "", -1)
		log.Debugf("Found version tag: %s, got version: %s", t.Name(), version)

		if !semver.IsValid(version) {
			log.Debugf("Not a semver tag: %s.", version)
			log.Debugf("Ignoring..")
			return nil
		}

		versions = append(versions, version)

		switch semver.Compare(version, latestVersion) {
		case 1:
			latestVersion = version
			return nil
		case 0, -1:
			return nil
		default:
			return nil
		}
	})

	if err != nil {
		return nil, "", err
	}

	return versions, latestVersion, nil
}

func decomposeURI(uri string) (string, string, string, error) {
	const gitURIPattern = `git@.+(\.git)?\?`
	const currentRefPattern = `ref=(.+)$`
	const repoNamePattern = `/([^/]+)(\.git)?[\?|$]`

	gitURIRe := regexp.MustCompile(gitURIPattern)
	currentRefRe := regexp.MustCompile(currentRefPattern)
	repoNameRe := regexp.MustCompile(repoNamePattern)

	bareGitURI := gitURIRe.FindAllString(uri, 1)
	bareCurrentRef := currentRefRe.FindAllStringSubmatch(uri, 1)
	bareRepoName := repoNameRe.FindAllStringSubmatch(uri, 1)

	if bareGitURI == nil || len(bareGitURI) != 1 {
		return "", "", "", fmt.Errorf("invalid submodule uri encountered: %s", uri)
	}

	if bareCurrentRef == nil {
		return "", "", "", fmt.Errorf("invalid current ref encountered: %s", uri)
	}

	if bareRepoName == nil {
		return "", "", "", fmt.Errorf("invalid repo name encountered: %s", uri)
	}

	gitURI := strings.Replace(bareGitURI[0], "/", ":", 1)
	gitURI = strings.Replace(gitURI, "?", "", 1)

	currentRef := bareCurrentRef[0][1]
	repoName := bareRepoName[0][1]

	return gitURI, currentRef, repoName, nil
}

// EvaluateGitModule : Extract module information from a git-hosted terraform
//                     module.
func EvaluateGitModule(uri string, sshKeyFile string) (*internals.Module, error) {
	gitURI, currentRef, repoName, err := decomposeURI(uri)

	if err != nil {
		return nil, err
	}

	if !semver.IsValid(currentRef) {
		log.Fatal("Invalid SemVer tag in module source string: ", currentRef)
	}

	log.Debugf("Extracting latest version tag from %s, current version: %s",
		repoName, currentRef)

	auth, err := ssh.NewPublicKeysFromFile("git", sshKeyFile, "")

	if err != nil {
		return nil, err
	}

	repo, clonePath, err := cloneGitRepo(gitURI, auth)

	if err != nil {
		return nil, err
	}

	versions, latestVersion, err := getVersions(repo)

	if err != nil {
		return nil, err
	}

	w, err := repo.Worktree()

	if err != nil {
		return nil, err
	}

	w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewTagReferenceName(currentRef),
	})

	return &internals.Module{
		Dependency: internals.Dependency{
			CurrentVersion: currentRef,
			LatestVersion:  latestVersion,
			Name:           repoName,
			Versions:       versions,
		},
		Source: gitURI,
		Path:   clonePath,
	}, nil
}
