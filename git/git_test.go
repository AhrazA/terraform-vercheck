package git

import (
	"github.com/go-git/go-git/v5"
	"testing"
)

func TestGitUriDecompose(t *testing.T) {
	tests := []struct {
		source     string
		gitURI     string
		currentRef string
		repoName   string
	}{
		{
			"git::ssh://git@github.com/AhrazA/somerepo.git?ref=v3.2.0",
			"git@github.com:AhrazA/somerepo.git",
			"v3.2.0",
			"somerepo.git",
		},
		{
			"git::ssh://git@github.com/AhrazA/somerepo.git?ref=v3.1.0",
			"git@github.com:AhrazA/somerepo.git",
			"v3.1.0",
			"somerepo.git",
		},
		{
			"git::ssh://git@github.com/AhrazA/somerepo?ref=v3.0.0",
			"git@github.com:AhrazA/somerepo",
			"v3.0.0",
			"somerepo",
		},
	}

	for _, test := range tests {
		gitURI, currentRef, repoName, err := decomposeURI(test.source)

		if err != nil {
			t.Error(err)
		}

		if gitURI != test.gitURI || currentRef != test.currentRef || repoName != test.repoName {
			t.Errorf("Expected URI:\n\t%s\nGot URI:\n\t%s\n"+
				"Expected ref:\n\t%s\nGot ref:\n\t%s\n"+
				"Expected name:\n\t%s\nGot name:\n\t%s\n",
				test.gitURI, gitURI,
				test.currentRef, currentRef,
				test.repoName, repoName)
		}
	}
}

func TestGetVersions(t *testing.T) {
	url := "https://github.com/helm/helm.git"
	repo, err := git.PlainClone(t.TempDir(), false, &git.CloneOptions{
		URL: url,
	})

	if err != nil {
		t.Fatal(err)
	}

	versions, latestVersion, err := getVersions(repo)

	if err != nil {
		t.Fatal(err)
	}

	if latestVersion == "v0.0.0" || len(versions) <= 1 {
		t.Errorf("Failed to parse versions on repo:\n\t%s", url)
	}
}

// TODO
func TestEvaluateGitModule(t *testing.T) {
}
