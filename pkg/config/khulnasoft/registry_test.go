package khulnasoft_test

import (
	"testing"

	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config/khulnasoft"
)

func TestRegistry_Validate(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title    string
		registry *khulnasoft.Registry
		isErr    bool
	}{
		{
			title: "github_content",
			registry: &khulnasoft.Registry{
				RepoOwner: "khulnasoftproj",
				RepoName:  "khulnasoft-registry",
				Ref:       "v0.8.0",
				Path:      "foo.yaml",
				Type:      "github_content",
			},
		},
		{
			title: "github_content repo_owner is required",
			registry: &khulnasoft.Registry{
				RepoName: "khulnasoft-registry",
				Ref:      "v0.8.0",
				Path:     "foo.yaml",
				Type:     "github_content",
			},
			isErr: true,
		},
		{
			title: "github_content repo_name is required",
			registry: &khulnasoft.Registry{
				RepoOwner: "khulnasoftproj",
				Ref:       "v0.8.0",
				Path:      "foo.yaml",
				Type:      "github_content",
			},
			isErr: true,
		},
		{
			title: "github_content ref is required",
			registry: &khulnasoft.Registry{
				RepoOwner: "khulnasoftproj",
				RepoName:  "khulnasoft-registry",
				Path:      "foo.yaml",
				Type:      "github_content",
			},
			isErr: true,
		},
		{
			title: "local",
			registry: &khulnasoft.Registry{
				Path: "foo.yaml",
				Type: "local",
			},
		},
		{
			title: "local path is required",
			registry: &khulnasoft.Registry{
				Type: "local",
			},
			isErr: true,
		},
		{
			title: "invalid type",
			registry: &khulnasoft.Registry{
				Type: "invalid-type",
			},
			isErr: true,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if err := d.registry.Validate(); err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
		})
	}
}

func TestRegistry_GetFilePath(t *testing.T) {
	t.Parallel()
	data := []struct {
		title       string
		exp         string
		registry    *khulnasoft.Registry
		rootDir     string
		homeDir     string
		cfgFilePath string
		isErr       bool
	}{
		{
			title:       "normal",
			exp:         "ci/foo.yaml",
			rootDir:     "/root/.khulnasoft",
			homeDir:     "/root",
			cfgFilePath: "ci/khulnasoft.yaml",
			registry: &khulnasoft.Registry{
				Path: "foo.yaml",
				Type: "local",
			},
		},
		{
			title:   "github_content",
			exp:     "/root/.khulnasoft/registries/github_content/github.com/khulnasoftproj/khulnasoft-registry/v0.8.0/foo.yaml",
			rootDir: "/root/.khulnasoft",
			registry: &khulnasoft.Registry{
				RepoOwner: "khulnasoftproj",
				RepoName:  "khulnasoft-registry",
				Ref:       "v0.8.0",
				Path:      "foo.yaml",
				Type:      "github_content",
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			p, err := d.registry.GetFilePath(d.rootDir, d.cfgFilePath)
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
			if p != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, p)
			}
		})
	}
}
