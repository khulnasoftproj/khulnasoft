package khulnasoft_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config/khulnasoft"
	"gopkg.in/yaml.v2"
)

func TestConfig_UnmarshalYAML(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title string
		yaml  string
		exp   *khulnasoft.Config
	}{
		{
			title: "standard registry",
			yaml: `
registries:
- type: standard
  ref: v0.2.0
packages:
- name: cmdx
  registry: standard
  version: v1.6.0
`,
			exp: &khulnasoft.Config{
				Registries: khulnasoft.Registries{
					"standard": &khulnasoft.Registry{
						Name:      "standard",
						RepoOwner: "khulnasoftproj",
						RepoName:  "khulnasoft-registry",
						Path:      "registry.yaml",
						Type:      "github_content",
						Ref:       "v0.2.0",
					},
				},
				Packages: []*khulnasoft.Package{
					{
						Name:     "cmdx",
						Registry: "standard",
						Version:  "v1.6.0",
					},
				},
			},
		},
		{
			title: "parse package name with version",
			yaml: `
registries:
- type: standard
  ref: v0.2.0
packages:
- name: sulaiman-coder/cmdx@v1.6.0
  registry: standard
`,
			exp: &khulnasoft.Config{
				Registries: khulnasoft.Registries{
					"standard": &khulnasoft.Registry{
						Name:      "standard",
						Type:      "github_content",
						RepoOwner: "khulnasoftproj",
						RepoName:  "khulnasoft-registry",
						Path:      "registry.yaml",
						Ref:       "v0.2.0",
					},
				},
				Packages: []*khulnasoft.Package{
					{
						Name:     "sulaiman-coder/cmdx",
						Registry: "standard",
						Version:  "v1.6.0",
					},
				},
			},
		},
	}

	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			cfg := &khulnasoft.Config{}
			if err := yaml.Unmarshal([]byte(d.yaml), cfg); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(d.exp, cfg); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
