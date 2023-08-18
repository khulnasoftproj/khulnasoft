package exec_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/khulnasoftproj/khulnasoft/v2/pkg/checksum"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config"
	finder "github.com/khulnasoftproj/khulnasoft/v2/pkg/config-finder"
	reader "github.com/khulnasoftproj/khulnasoft/v2/pkg/config-reader"
	execCtrl "github.com/khulnasoftproj/khulnasoft/v2/pkg/controller/exec"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/controller/which"
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
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
)

func Test_controller_Exec(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name    string
		files   map[string]string
		dirs    []string
		links   map[string]string
		env     map[string]string
		param   *config.Param
		exeName string
		rt      *runtime.Runtime
		args    []string
		isErr   bool
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
			dirs: []string{
				"/home/foo/workspace/.git",
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
				"/home/foo/.local/share/khulnasoftproj-khulnasoft/pkgs/github_content/github.com/khulnasoftproj/khulnasoft-installer/v1.0.0/khulnasoft-installer/khulnasoft-installer": "",
				"/home/foo/workspace/khulnasoft-policy.yaml": `
registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- type: local
`,
				"/home/foo/.local/share/khulnasoftproj-khulnasoft/policies/home/foo/workspace/khulnasoft-policy.yaml": `
registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- type: local
`,
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
		},
	}
	logE := logrus.NewEntry(logrus.New())
	ctx := context.Background()
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
			ghDownloader := download.NewGitHubContentFileDownloader(nil, download.NewHTTPDownloader(http.DefaultClient))
			osEnv := osenv.NewMock(d.env)
			whichCtrl := which.New(d.param, finder.NewConfigFinder(fs), reader.New(fs, d.param), registry.New(d.param, ghDownloader, fs, d.rt, &cosign.MockVerifier{}, &slsa.MockVerifier{}), d.rt, osEnv, fs, linker)
			downloader := download.NewDownloader(nil, download.NewHTTPDownloader(http.DefaultClient))
			executor := &exec.Mock{}
			pkgInstaller := installpackage.New(d.param, downloader, d.rt, fs, linker, nil, &checksum.Calculator{}, unarchive.New(executor, fs), &policy.Checker{}, &cosign.MockVerifier{}, &slsa.MockVerifier{}, &installpackage.MockGoInstallInstaller{}, &installpackage.MockGoBuildInstaller{}, &installpackage.MockCargoPackageInstaller{})
			policyFinder := policy.NewConfigFinder(fs)
			ctrl := execCtrl.New(d.param, pkgInstaller, whichCtrl, executor, osEnv, fs, policy.NewReader(fs, policy.NewValidator(d.param, fs), policyFinder, policy.NewConfigReader(fs)), policyFinder)
			if err := ctrl.Exec(ctx, logE, d.param, d.exeName, d.args...); err != nil {
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

func downloadTestFile(uri, tempDir string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, uri, nil) //nolint:noctx
	if err != nil {
		return "", fmt.Errorf("create a request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send a HTTP request: %w", err)
	}
	defer resp.Body.Close()
	filePath := filepath.Join(tempDir, "registry.yaml")
	f, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("create a file: %w", err)
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		return "", fmt.Errorf("write a response body to a file: %w", err)
	}
	return filePath, nil
}

func Benchmark_controller_Exec(b *testing.B) { //nolint:funlen,gocognit
	data := []struct {
		name    string
		files   map[string]string
		links   map[string]string
		env     map[string]string
		param   *config.Param
		exeName string
		rt      *runtime.Runtime
		args    []string
		isErr   bool
	}{
		{
			name: "normal",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			param: &config.Param{
				PWD:            "/home/foo/workspace",
				RootDir:        "/home/foo/.local/share/khulnasoftproj-khulnasoft",
				MaxParallelism: 5,
			},
			exeName: "khulnasoft-installer",
			files:   map[string]string{},
		},
	}
	logE := logrus.NewEntry(logrus.New())
	ctx := context.Background()
	for _, d := range data {
		d := d
		b.Run("normal", func(b *testing.B) {
			tempDir := b.TempDir()
			d.param.ConfigFilePath = filepath.Join(tempDir, "khulnasoft.yaml")
			d.files[d.param.ConfigFilePath] = `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- name: khulnasoftproj/khulnasoft-installer@v1.0.0
`
			if _, err := downloadTestFile("https://raw.githubusercontent.com/khulnasoftproj/khulnasoft-registry/v2.19.0/registry.yaml", tempDir); err != nil {
				b.Fatal(err)
			}
			fs, err := testutil.NewFs(d.files)
			if err != nil {
				b.Fatal(err)
			}
			linker := domain.NewMockLinker(fs)
			for dest, src := range d.links {
				if err := linker.Symlink(dest, src); err != nil {
					b.Fatal(err)
				}
			}
			ghDownloader := download.NewGitHubContentFileDownloader(nil, download.NewHTTPDownloader(http.DefaultClient))
			osEnv := osenv.NewMock(d.env)
			whichCtrl := which.New(d.param, finder.NewConfigFinder(fs), reader.New(fs, d.param), registry.New(d.param, ghDownloader, afero.NewOsFs(), d.rt, &cosign.MockVerifier{}, &slsa.MockVerifier{}), d.rt, osEnv, fs, linker)
			downloader := download.NewDownloader(nil, download.NewHTTPDownloader(http.DefaultClient))
			executor := &exec.Mock{}
			pkgInstaller := installpackage.New(d.param, downloader, d.rt, fs, linker, nil, &checksum.Calculator{}, unarchive.New(executor, fs), &policy.Checker{}, &cosign.MockVerifier{}, &slsa.MockVerifier{}, &installpackage.MockGoInstallInstaller{}, &installpackage.MockGoBuildInstaller{}, &installpackage.MockCargoPackageInstaller{})
			ctrl := execCtrl.New(d.param, pkgInstaller, whichCtrl, executor, osEnv, fs, &policy.MockReader{}, policy.NewConfigFinder(fs))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				func() {
					if err := ctrl.Exec(ctx, logE, d.param, d.exeName, d.args...); err != nil {
						if d.isErr {
							return
						}
						b.Fatal(err)
					}
					if d.isErr {
						b.Fatal("error must be returned")
					}
				}()
			}
		})
	}
}
