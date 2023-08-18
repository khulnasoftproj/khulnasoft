package install_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/khulnasoftproj/khulnasoft/v2/pkg/checksum"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config"
	finder "github.com/khulnasoftproj/khulnasoft/v2/pkg/config-finder"
	reader "github.com/khulnasoftproj/khulnasoft/v2/pkg/config-reader"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/controller/install"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/cosign"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/domain"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/download"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/exec"
	registry "github.com/khulnasoftproj/khulnasoft/v2/pkg/install-registry"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/installpackage"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/policy"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/runtime"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/slsa"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/testutil"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/unarchive"
	"github.com/sirupsen/logrus"
)

func TestController_Install(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name              string
		files             map[string]string
		dirs              []string
		links             map[string]string
		param             *config.Param
		rt                *runtime.Runtime
		registryInstaller registry.Installer
		isErr             bool
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
			files: map[string]string{
				"/home/foo/workspace/khulnasoft.yaml": `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- name: khulnasoftproj/khulnasoft-installer@v1.0.0
`,
				"/home/foo/workspace/khulnasoft-policy.yaml": `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- registry: standard
`,
				"/home/foo/workspace/registry.yaml": `packages:
- type: github_content
  repo_owner: khulnasoftproj
  repo_name: khulnasoft-installer
  path: khulnasoft-installer
`,
				"/home/foo/.local/share/khulnasoftproj-khulnasoft/pkgs/github_content/github.com/khulnasoftproj/khulnasoft-installer/v1.0.0/khulnasoft-installer/khulnasoft-installer":                                                       ``,
				fmt.Sprintf("/home/foo/.local/share/khulnasoftproj-khulnasoft/internal/pkgs/github_release/github.com/khulnasoftproj/khulnasoft-proxy/%s/khulnasoft-proxy_linux_amd64.tar.gz/khulnasoft-proxy", installpackage.ProxyVersion): ``,
				"/home/foo/.local/share/khulnasoftproj-khulnasoft/bin/khulnasoft-installer": ``,
				"/home/foo/.local/share/khulnasoftproj-khulnasoft/khulnasoft-proxy":         ``,
			},
			dirs: []string{
				"/home/foo/workspace/.git",
			},
			links: map[string]string{
				"../khulnasoft-proxy": "/home/foo/.local/share/khulnasoftproj-khulnasoft/bin/khulnasoft-installer",
				fmt.Sprintf("../internal/pkgs/github_release/github.com/khulnasoftproj/khulnasoft-proxy/%s/khulnasoft-proxy_linux_amd64.tar.gz/khulnasoft-proxy", installpackage.ProxyVersion): "/home/foo/.local/share/khulnasoftproj-khulnasoft/bin/khulnasoft-proxy",
			},
		},
	}
	logE := logrus.NewEntry(logrus.New())
	ctx := context.Background()
	registryDownloader := download.NewGitHubContentFileDownloader(nil, download.NewHTTPDownloader(http.DefaultClient))
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			fs, err := testutil.NewFs(d.files, d.dirs...)
			if err != nil {
				t.Fatal(err)
			}
			linker := domain.NewMockLinker(fs)
			for dest, src := range d.links {
				if err := linker.Symlink(dest, src); err != nil {
					t.Fatal(err)
				}
			}
			downloader := download.NewDownloader(nil, download.NewHTTPDownloader(http.DefaultClient))
			executor := &exec.Mock{}
			pkgInstaller := installpackage.New(d.param, downloader, d.rt, fs, linker, nil, &checksum.Calculator{}, unarchive.New(executor, fs), &policy.Checker{}, &cosign.MockVerifier{}, &slsa.MockVerifier{}, &installpackage.MockGoInstallInstaller{}, &installpackage.MockGoBuildInstaller{}, &installpackage.MockCargoPackageInstaller{})
			policyFinder := policy.NewConfigFinder(fs)
			policyReader := policy.NewReader(fs, &policy.MockValidator{}, policyFinder, policy.NewConfigReader(fs))
			ctrl := install.New(d.param, finder.NewConfigFinder(fs), reader.New(fs, d.param), registry.New(d.param, registryDownloader, fs, d.rt, &cosign.MockVerifier{}, &slsa.MockVerifier{}), pkgInstaller, fs, d.rt, policyReader, policyFinder)
			if err := ctrl.Install(ctx, logE, d.param); err != nil {
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
