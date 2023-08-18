package policy_test

import (
	"testing"

	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config/khulnasoft"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config/registry"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/policy"
	"github.com/sirupsen/logrus"
)

func TestChecker_ValidatePackage(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name     string
		isErr    bool
		pkg      *config.Package
		policies []*policy.Config
	}{
		{
			name: "no policy",
			pkg: &config.Package{
				Package: &khulnasoft.Package{
					Name:    "sulaiman-coder/tfcmt",
					Version: "v4.0.0",
				},
				PackageInfo: &registry.PackageInfo{},
				Registry: &khulnasoft.Registry{
					Type:      "github_content",
					Name:      registryTypeStandard,
					RepoOwner: "khulnasoftproj",
					RepoName:  "khulnasoft-registry",
					Path:      "registry.yaml",
					Ref:       "v3.90.0",
				},
			},
		},
		{
			name: "normal",
			pkg: &config.Package{
				Package: &khulnasoft.Package{
					Name:    "sulaiman-coder/tfcmt",
					Version: "v4.0.0",
				},
				PackageInfo: &registry.PackageInfo{},
				Registry: &khulnasoft.Registry{
					Type:      "github_content",
					Name:      registryTypeStandard,
					RepoOwner: "khulnasoftproj",
					RepoName:  "khulnasoft",
					Path:      "registry.yaml",
					Ref:       "v1.90.0",
				},
			},
			policies: []*policy.Config{
				{
					YAML: &policy.ConfigYAML{
						Packages: []*policy.Package{
							{
								Name: "cli/cli",
							},
							{
								Name:         "sulaiman-coder/tfcmt",
								Version:      `semver(">= 3.0.0")`,
								RegistryName: "standard",
								Registry: &policy.Registry{
									Type:      "github_content",
									Name:      registryTypeStandard,
									RepoOwner: "khulnasoftproj",
									RepoName:  "khulnasoft",
									Path:      "registry.yaml",
								},
							},
						},
					},
				},
			},
		},
	}
	checker := &policy.Checker{}
	logE := logrus.NewEntry(logrus.New())
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			if err := checker.ValidatePackage(logE, d.pkg, d.policies); err != nil {
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
