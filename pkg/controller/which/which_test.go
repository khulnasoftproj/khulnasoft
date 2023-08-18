package which_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config"
	finder "github.com/khulnasoftproj/khulnasoft/v2/pkg/config-finder"
	reader "github.com/khulnasoftproj/khulnasoft/v2/pkg/config-reader"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config/khulnasoft"
	cfgRegistry "github.com/khulnasoftproj/khulnasoft/v2/pkg/config/registry"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/controller/which"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/cosign"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/domain"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/download"
	registry "github.com/khulnasoftproj/khulnasoft/v2/pkg/install-registry"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/runtime"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/slsa"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/testutil"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
)

func stringP(s string) *string {
	return &s
}

func Test_controller_Which(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name    string
		files   map[string]string
		links   map[string]string
		env     map[string]string
		param   *config.Param
		exeName string
		rt      *runtime.Runtime
		isErr   bool
		exp     *which.FindResult
	}{
		{
			name: "normal",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			param: &config.Param{
				PWD:            "/home/foo/workspace",
				ConfigFilePath: "khulnasoft.yaml",
				RootDir:        "/home/foo/.local/share/khulnasoftproj-khulnasoft",
				MaxParallelism: 5,
			},
			exeName: "khulnasoft-installer",
			files: map[string]string{
				"/home/foo/workspace/khulnasoft.yaml": `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- name: khulnasoftproj/khulnasoft-installer@v1.0.0
`,
				"/home/foo/workspace/registry.yaml": `packages:
- type: github_content
  repo_owner: khulnasoftproj
  repo_name: khulnasoft-installer
  path: khulnasoft-installer
`,
			},
			exp: &which.FindResult{
				Package: &config.Package{
					Package: &khulnasoft.Package{
						Name:     "khulnasoftproj/khulnasoft-installer",
						Registry: "standard",
						Version:  "v1.0.0",
					},
					PackageInfo: &cfgRegistry.PackageInfo{
						Type:      "github_content",
						RepoOwner: "khulnasoftproj",
						RepoName:  "khulnasoft-installer",
						Path:      stringP("khulnasoft-installer"),
					},
					Registry: &khulnasoft.Registry{
						Name: "standard",
						Type: "local",
						Path: "/home/foo/workspace/registry.yaml",
					},
				},
				File: &cfgRegistry.File{
					Name: "khulnasoft-installer",
				},
				Config: &khulnasoft.Config{
					Packages: []*khulnasoft.Package{
						{
							Name:     "khulnasoftproj/khulnasoft-installer",
							Registry: "standard",
							Version:  "v1.0.0",
						},
					},
					Registries: khulnasoft.Registries{
						"standard": {
							Name: "standard",
							Type: "local",
							Path: "/home/foo/workspace/registry.yaml",
						},
					},
				},

				ExePath:        "/home/foo/.local/share/khulnasoftproj-khulnasoft/pkgs/github_content/github.com/khulnasoftproj/khulnasoft-installer/v1.0.0/khulnasoft-installer/khulnasoft-installer",
				ConfigFilePath: "/home/foo/workspace/khulnasoft.yaml",
			},
		},
		{
			name: "outside khulnasoft",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			param: &config.Param{
				PWD:            "/home/foo/workspace",
				ConfigFilePath: "khulnasoft.yaml",
				RootDir:        "/home/foo/.local/share/khulnasoftproj-khulnasoft",
				MaxParallelism: 5,
			},
			exeName: "gh",
			env: map[string]string{
				"PATH": "/home/foo/.local/share/khulnasoftproj-khulnasoft/bin:/usr/local/bin:/usr/bin",
			},
			files: map[string]string{
				"/home/foo/workspace/khulnasoft.yaml": `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- name: khulnasoftproj/khulnasoft-installer@v1.0.0
`,
				"/home/foo/workspace/registry.yaml": `packages:
- type: github_content
  repo_owner: khulnasoftproj
  repo_name: khulnasoft-installer
  path: khulnasoft-installer
`,
				"/usr/local/foo/gh": "",
			},
			links: map[string]string{
				"../foo/gh": "/usr/local/bin/gh",
			},
			exp: &which.FindResult{
				ExePath: "/usr/local/bin/gh",
			},
		},
		{
			name: "global config",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			param: &config.Param{
				PWD:                   "/home/foo/workspace",
				RootDir:               "/home/foo/.local/share/khulnasoftproj-khulnasoft",
				MaxParallelism:        5,
				GlobalConfigFilePaths: []string{"/etc/khulnasoft/khulnasoft.yaml"},
			},
			exeName: "khulnasoft-installer",
			files: map[string]string{
				"/etc/khulnasoft/khulnasoft.yaml": `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- name: sulaiman-coder/ci-info@v1.0.0
- name: khulnasoftproj/khulnasoft-installer@v1.0.0
`,
				"/etc/khulnasoft/registry.yaml": `packages:
- type: github_release
  repo_owner: sulaiman-coder
  repo_name: ci-info
  asset: "ci-info_{{.Arch}}-{{.OS}}.tar.gz"
- type: github_release
  repo_owner: sulaiman-coder
  repo_name: github-comment
  asset: "github-comment_{{.Arch}}-{{.OS}}.tar.gz"
- type: github_content
  repo_owner: khulnasoftproj
  repo_name: khulnasoft-installer
  path: khulnasoft-installer
`,
			},
			exp: &which.FindResult{
				Package: &config.Package{
					Package: &khulnasoft.Package{
						Name:     "khulnasoftproj/khulnasoft-installer",
						Registry: "standard",
						Version:  "v1.0.0",
					},
					PackageInfo: &cfgRegistry.PackageInfo{
						Type:      "github_content",
						RepoOwner: "khulnasoftproj",
						RepoName:  "khulnasoft-installer",
						Path:      stringP("khulnasoft-installer"),
					},
					Registry: &khulnasoft.Registry{
						Name: "standard",
						Type: "local",
						Path: "/etc/khulnasoft/registry.yaml",
					},
				},
				File: &cfgRegistry.File{
					Name: "khulnasoft-installer",
				},
				Config: &khulnasoft.Config{
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
					Registries: khulnasoft.Registries{
						"standard": {
							Name: "standard",
							Type: "local",
							Path: "/etc/khulnasoft/registry.yaml",
						},
					},
				},
				ExePath:        "/home/foo/.local/share/khulnasoftproj-khulnasoft/pkgs/github_content/github.com/khulnasoftproj/khulnasoft-installer/v1.0.0/khulnasoft-installer/khulnasoft-installer",
				ConfigFilePath: "/etc/khulnasoft/khulnasoft.yaml",
			},
		},
	}
	logE := logrus.NewEntry(logrus.New())
	ctx := context.Background()
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			fs, err := testutil.NewFs(d.files)
			if err != nil {
				t.Fatal(err)
			}
			linker := domain.NewMockLinker(fs)
			for dest, src := range d.links {
				if err := linker.Symlink(dest, src); err != nil {
					t.Fatal(err)
				}
			}
			downloader := download.NewGitHubContentFileDownloader(nil, download.NewHTTPDownloader(http.DefaultClient))
			ctrl := which.New(d.param, finder.NewConfigFinder(fs), reader.New(fs, d.param), registry.New(d.param, downloader, fs, d.rt, &cosign.MockVerifier{}, &slsa.MockVerifier{}), d.rt, osenv.NewMock(d.env), fs, linker)
			which, err := ctrl.Which(ctx, logE, d.param, d.exeName)
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
			if diff := cmp.Diff(d.exp, which); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
