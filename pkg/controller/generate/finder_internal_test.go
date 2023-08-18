package generate

import (
	"testing"

	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config/registry"
)

func Test_find(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name string
		pkg  *FindingPackage
		exp  string
	}{
		{
			name: "normal",
			pkg: &FindingPackage{
				PackageInfo: &registry.PackageInfo{
					RepoOwner: "sulaiman-coder",
					RepoName:  "ci-info",
				},
				RegistryName: registryStandard,
			},
			exp: "sulaiman-coder/ci-info",
		},
		{
			name: "search words",
			pkg: &FindingPackage{
				PackageInfo: &registry.PackageInfo{
					RepoOwner:   "sulaiman-coder",
					RepoName:    "ci-info",
					SearchWords: []string{"pull request"},
				},
				RegistryName: registryStandard,
			},
			exp: "sulaiman-coder/ci-info: pull request",
		},
		{
			name: "search words, registry",
			pkg: &FindingPackage{
				PackageInfo: &registry.PackageInfo{
					RepoOwner:   "sulaiman-coder",
					RepoName:    "ci-info",
					SearchWords: []string{"pull request"},
				},
				RegistryName: "foo",
			},
			exp: "sulaiman-coder/ci-info (foo): pull request",
		},
		{
			name: "search words, alias, registry",
			pkg: &FindingPackage{
				PackageInfo: &registry.PackageInfo{
					RepoOwner:   "sulaiman-coder",
					RepoName:    "ci-info",
					SearchWords: []string{"pull request"},
					Aliases: []*registry.Alias{
						{
							Name: "ci-info",
						},
					},
				},
				RegistryName: "foo",
			},
			exp: "sulaiman-coder/ci-info (ci-info) (foo): pull request",
		},
		{
			name: "search words, alias, command, registry",
			pkg: &FindingPackage{
				PackageInfo: &registry.PackageInfo{
					RepoOwner:   "sulaiman-coder",
					RepoName:    "ci-info",
					SearchWords: []string{"pull request"},
					Aliases: []*registry.Alias{
						{
							Name: "ci-info",
						},
					},
					Files: []*registry.File{
						{
							Name: "ci-info",
						},
						{
							Name: "ci",
						},
					},
				},
				RegistryName: "foo",
			},
			exp: "sulaiman-coder/ci-info (ci-info) (foo) [ci-info, ci]: pull request",
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			s := find(d.pkg)
			if s != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, s)
			}
		})
	}
}

func Test_getPreview(t *testing.T) {
	t.Parallel()
	data := []struct {
		name string
		pkg  *FindingPackage
		i    int
		w    int
		exp  string
	}{
		{
			name: "normal",
			pkg: &FindingPackage{
				PackageInfo: &registry.PackageInfo{
					RepoOwner:   "sulaiman-coder",
					RepoName:    "ci-info",
					Description: "CLI tool to get CI related information",
				},
				RegistryName: registryStandard,
			},
			w: 100,
			exp: `sulaiman-coder/ci-info

https://github.com/sulaiman-coder/ci-info
CLI tool to get CI related information`,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			s := getPreview(d.pkg, d.i, d.w)
			if s != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, s)
			}
		})
	}
}
