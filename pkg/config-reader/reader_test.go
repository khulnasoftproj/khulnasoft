package reader_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config"
	reader "github.com/khulnasoftproj/khulnasoft/v2/pkg/config-reader"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config/khulnasoft"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/testutil"
)

func Test_configReader_Read(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name           string
		exp            *khulnasoft.Config
		isErr          bool
		files          map[string]string
		configFilePath string
		homeDir        string
	}{
		{
			name:  "file isn't found",
			isErr: true,
		},
		{
			name: "normal",
			files: map[string]string{
				"/home/workspace/foo/khulnasoft.yaml": `registries:
- type: standard
  ref: v2.5.0
- type: local
  name: local
  path: registry.yaml
packages:`,
			},
			configFilePath: "/home/workspace/foo/khulnasoft.yaml",
			exp: &khulnasoft.Config{
				Registries: khulnasoft.Registries{
					"standard": {
						Type:      "github_content",
						Name:      "standard",
						Ref:       "v2.5.0",
						RepoOwner: "khulnasoftproj",
						RepoName:  "khulnasoft-registry",
						Path:      "registry.yaml",
					},
					"local": {
						Type: "local",
						Name: "local",
						Path: "/home/workspace/foo/registry.yaml",
					},
				},
				Packages: []*khulnasoft.Package{},
			},
		},
		{
			name: "import package",
			files: map[string]string{
				"/home/workspace/foo/khulnasoft.yaml": `registries:
- type: standard
  ref: v2.5.0
packages:
- name: sulaiman-coder/ci-info@v1.0.0
- import: khulnasoft-installer.yaml
`,
				"/home/workspace/foo/khulnasoft-installer.yaml": `packages:
- name: khulnasoftproj/khulnasoft-installer@v1.0.0
`,
			},
			configFilePath: "/home/workspace/foo/khulnasoft.yaml",
			exp: &khulnasoft.Config{
				Registries: khulnasoft.Registries{
					"standard": {
						Type:      "github_content",
						Name:      "standard",
						Ref:       "v2.5.0",
						RepoOwner: "khulnasoftproj",
						RepoName:  "khulnasoft-registry",
						Path:      "registry.yaml",
					},
				},
				Packages: []*khulnasoft.Package{
					{
						Name:     "sulaiman-coder/ci-info",
						Registry: "standard",
						Version:  "v1.0.0",
					},
					{
						Name:     "khulnasoftproj/khulnasoft-installer",
						Registry: "standard",
						Version:  "v1.0.0",
					},
				},
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			fs, err := testutil.NewFs(d.files)
			if err != nil {
				t.Fatal(err)
			}
			reader := reader.New(fs, &config.Param{
				HomeDir: d.homeDir,
			})
			cfg := &khulnasoft.Config{}
			if err := reader.Read(d.configFilePath, cfg); err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
			if diff := cmp.Diff(d.exp, cfg); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
