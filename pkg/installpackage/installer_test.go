package installpackage_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/khulnasoftproj/khulnasoft/v2/pkg/checksum"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config/khulnasoft"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config/registry"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/cosign"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/domain"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/download"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/installpackage"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/policy"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/runtime"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/slsa"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/testutil"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/unarchive"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func stringP(s string) *string {
	return &s
}

func Test_installer_InstallPackages(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name       string
		files      map[string]string
		links      map[string]string
		param      *config.Param
		rt         *runtime.Runtime
		cfg        *khulnasoft.Config
		registries map[string]*registry.Config
		executor   installpackage.Executor
		binDir     string
		isErr      bool
	}{
		{
			name: "file already exists",
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
			cfg: &khulnasoft.Config{
				Registries: khulnasoft.Registries{
					"standard": {
						Name:      "standard",
						Type:      "github_content",
						RepoOwner: "khulnasoftproj",
						RepoName:  "khulnasoft-registry",
						Ref:       "v2.15.0",
						Path:      "registry.yaml",
					},
				},
				Packages: []*khulnasoft.Package{
					{
						Name:     "sulaiman-coder/ci-info",
						Registry: "standard",
						Version:  "v2.0.3",
					},
				},
			},
			registries: map[string]*registry.Config{
				"standard": {
					PackageInfos: registry.PackageInfos{
						{
							Type:      "github_release",
							RepoOwner: "sulaiman-coder",
							RepoName:  "ci-info",
							Asset:     stringP("ci-info_{{trimV .Version}}_{{.OS}}_amd64.tar.gz"),
						},
					},
				},
			},
			binDir: "/home/foo/.local/share/khulnasoftproj-khulnasoft/bin",
			files: map[string]string{
				"/home/foo/.local/share/khulnasoftproj-khulnasoft/pkgs/github_release/github.com/sulaiman-coder/ci-info/v2.0.3/ci-info_2.0.3_linux_amd64.tar.gz/ci-info": ``,
			},
			links: map[string]string{
				"../khulnasoft-proxy": "/home/foo/.local/share/khulnasoftproj-khulnasoft/bin/ci-info",
			},
		},
		{
			name: "only link",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			param: &config.Param{
				PWD:            "/home/foo/workspace",
				ConfigFilePath: "khulnasoft.yaml",
				RootDir:        "/home/foo/.local/share/khulnasoftproj-khulnasoft",
				MaxParallelism: 5,
				OnlyLink:       true,
			},
			cfg: &khulnasoft.Config{
				Registries: khulnasoft.Registries{
					"standard": {
						Name:      "standard",
						Type:      "github_content",
						RepoOwner: "khulnasoftproj",
						RepoName:  "khulnasoft-registry",
						Ref:       "v2.15.0",
						Path:      "registry.yaml",
					},
				},
				Packages: []*khulnasoft.Package{
					{
						Name:     "sulaiman-coder/ci-info",
						Registry: "standard",
						Version:  "v2.0.3",
					},
				},
			},
			registries: map[string]*registry.Config{
				"standard": {
					PackageInfos: registry.PackageInfos{
						{
							Type:      "github_release",
							RepoOwner: "sulaiman-coder",
							RepoName:  "ci-info",
							Asset:     stringP("ci-info_{{trimV .Version}}_{{.OS}}_amd64.tar.gz"),
						},
					},
				},
			},
			binDir: "/home/foo/.local/share/khulnasoftproj-khulnasoft/bin",
			files: map[string]string{
				"/home/foo/.local/share/khulnasoftproj-khulnasoft/pkgs/github_release/github.com/sulaiman-coder/ci-info/v2.0.3/ci-info_2.0.3_linux_amd64.tar.gz/ci-info": ``,
			},
			links: map[string]string{
				"../khulnasoft-proxy": "/home/foo/.local/share/khulnasoftproj-khulnasoft/bin/ci-info",
			},
		},
		{
			name: "no package",
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
			cfg: &khulnasoft.Config{
				Registries: khulnasoft.Registries{
					"standard": {
						Name:      "standard",
						Type:      "github_content",
						RepoOwner: "khulnasoftproj",
						RepoName:  "khulnasoft-registry",
						Ref:       "v2.15.0",
						Path:      "registry.yaml",
					},
				},
				Packages: []*khulnasoft.Package{},
			},
			registries: map[string]*registry.Config{
				"standard": {
					PackageInfos: registry.PackageInfos{},
				},
			},
			binDir: "/home/foo/.local/share/khulnasoftproj-khulnasoft/bin",
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
			linker := domain.NewMockLinker(afero.NewMemMapFs())
			for dest, src := range d.links {
				if err := linker.Symlink(dest, src); err != nil {
					t.Fatal(err)
				}
			}
			downloader := download.NewDownloader(nil, download.NewHTTPDownloader(http.DefaultClient))
			ctrl := installpackage.New(d.param, downloader, d.rt, fs, linker, nil, &checksum.Calculator{}, unarchive.New(d.executor, fs), &policy.Checker{}, &cosign.MockVerifier{}, &slsa.MockVerifier{}, &installpackage.MockGoInstallInstaller{}, &installpackage.MockGoBuildInstaller{}, &installpackage.MockCargoPackageInstaller{})
			if err := ctrl.InstallPackages(ctx, logE, &installpackage.ParamInstallPackages{
				Config:         d.cfg,
				Registries:     d.registries,
				ConfigFilePath: d.param.ConfigFilePath,
			}); err != nil {
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

func Test_installer_InstallPackage(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name     string
		files    map[string]string
		param    *config.Param
		rt       *runtime.Runtime
		pkg      *config.Package
		executor installpackage.Executor
		isTest   bool
		isErr    bool
	}{
		{
			name: "file already exists",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:      "github_release",
					RepoOwner: "sulaiman-coder",
					RepoName:  "ci-info",
					Asset:     stringP("ci-info_{{trimV .Version}}_{{.OS}}_amd64.tar.gz"),
				},
				Package: &khulnasoft.Package{
					Name:     "sulaiman-coder/ci-info",
					Registry: "standard",
					Version:  "v2.0.3",
				},
				Registry: &khulnasoft.Registry{
					Name:      "standard",
					Type:      "github_content",
					RepoOwner: "khulnasoftproj",
					RepoName:  "khulnasoft-registry",
					Ref:       "v2.15.0",
					Path:      "registry.yaml",
				},
			},
			param: &config.Param{
				RootDir: "/home/foo/.local/share/khulnasoftproj-khulnasoft",
			},
			files: map[string]string{
				"/home/foo/.local/share/khulnasoftproj-khulnasoft/pkgs/github_release/github.com/sulaiman-coder/ci-info/v2.0.3/ci-info_2.0.3_linux_amd64.tar.gz/ci-info": ``,
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
			downloader := download.NewDownloader(nil, download.NewHTTPDownloader(http.DefaultClient))
			ctrl := installpackage.New(d.param, downloader, d.rt, fs, nil, nil, &checksum.Calculator{}, unarchive.New(d.executor, fs), &policy.Checker{}, &cosign.MockVerifier{}, &slsa.MockVerifier{}, &installpackage.MockGoInstallInstaller{}, &installpackage.MockGoBuildInstaller{}, &installpackage.MockCargoPackageInstaller{})
			if err := ctrl.InstallPackage(ctx, logE, &installpackage.ParamInstallPackage{
				Pkg: d.pkg,
			}); err != nil {
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
