package config_test

import (
	"testing"

	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config/khulnasoft"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config/registry"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/runtime"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/util"
)

func TestPackage_RenderAsset(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title string
		exp   string
		pkg   *config.Package
		rt    *runtime.Runtime
	}{
		{
			title: "github_archive",
			exp:   "",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type: "github_archive",
				},
			},
		},
		{
			title: "github_content",
			exp:   "foo",
			pkg: &config.Package{
				Package: &khulnasoft.Package{
					Version: "v1.0.0",
				},
				PackageInfo: &registry.PackageInfo{
					Type: "github_content",
					Path: util.StrP("foo"),
				},
			},
		},
		{
			title: "github_release",
			exp:   "foo-1.0.0.zip",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:   "github_release",
					Asset:  util.StrP("foo-{{trimV .Version}}.{{.Format}}"),
					Format: "zip",
				},
				Package: &khulnasoft.Package{
					Version: "v1.0.0",
				},
			},
		},
		{
			title: "http",
			exp:   "foo-1.0.0.zip",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:   "http",
					URL:    util.StrP("https://example.com/foo-{{trimV .Version}}.{{.Format}}"),
					Format: "zip",
				},
				Package: &khulnasoft.Package{
					Version: "v1.0.0",
				},
			},
		},
		{
			title: "windows add .exe",
			exp:   "foo-windows-amd64.exe",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:   "github_release",
					Asset:  util.StrP("foo-{{.OS}}-{{.Arch}}"),
					Format: "raw",
				},
				Package: &khulnasoft.Package{
					Version: "v1.0.0",
				},
			},
			rt: &runtime.Runtime{
				GOOS:   "windows",
				GOARCH: "amd64",
			},
		},
		{
			title: "windows add .exe without Format",
			exp:   "foo-windows-amd64.exe",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:  "github_release",
					Asset: util.StrP("foo-{{.OS}}-{{.Arch}}"),
				},
				Package: &khulnasoft.Package{
					Version: "v1.0.0",
				},
			},
			rt: &runtime.Runtime{
				GOOS:   "windows",
				GOARCH: "amd64",
			},
		},
		{
			title: "windows",
			exp:   "foo-windows-amd64.tar.gz",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:  "github_release",
					Asset: util.StrP("foo-{{.OS}}-{{.Arch}}.tar.gz"),
				},
				Package: &khulnasoft.Package{
					Version: "v1.0.0",
				},
			},
			rt: &runtime.Runtime{
				GOOS:   "windows",
				GOARCH: "amd64",
			},
		},
	}
	rt := runtime.New()
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if d.rt == nil {
				d.rt = rt
			}
			asset, err := d.pkg.RenderAsset(d.rt)
			if err != nil {
				t.Fatal(err)
			}
			if asset != d.exp {
				t.Fatalf("wanted %v, got %v", d.exp, asset)
			}
		})
	}
}

func TestPackageInfo_GetPkgPath(t *testing.T) { //nolint:funlen
	t.Parallel()
	rootDir := "/tmp/khulnasoft"
	data := []struct {
		title string
		exp   string
		pkg   *config.Package
	}{
		{
			title: "github_archive",
			exp:   "/tmp/khulnasoft/pkgs/github_archive/github.com/tfutils/tfenv/v2.2.2",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:      "github_archive",
					RepoOwner: "tfutils",
					RepoName:  "tfenv",
				},
				Package: &khulnasoft.Package{
					Version: "v2.2.2",
				},
			},
		},
		{
			title: "github_content",
			exp:   "/tmp/khulnasoft/pkgs/github_content/github.com/khulnasoftproj/khulnasoft-installer/v0.2.0/khulnasoft-installer",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:      "github_content",
					Path:      util.StrP("khulnasoft-installer"),
					RepoOwner: "khulnasoftproj",
					RepoName:  "khulnasoft-installer",
				},
				Package: &khulnasoft.Package{
					Version: "v0.2.0",
				},
			},
		},
		{
			title: "github_release",
			exp:   "/tmp/khulnasoft/pkgs/github_release/github.com/sulaiman-coder/ci-info/v0.7.7/ci-info.tar.gz",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:      "github_release",
					RepoOwner: "sulaiman-coder",
					RepoName:  "ci-info",
					Asset:     util.StrP("ci-info.{{.Format}}"),
					Format:    "tar.gz",
				},
				Package: &khulnasoft.Package{
					Version: "v0.7.7",
				},
			},
		},
		{
			title: "http",
			exp:   "/tmp/khulnasoft/pkgs/http/example.com/foo-1.0.0.zip",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:   "http",
					URL:    util.StrP("https://example.com/foo-{{trimV .Version}}.{{.Format}}"),
					Format: "zip",
				},
				Package: &khulnasoft.Package{
					Version: "v1.0.0",
				},
			},
		},
	}
	rt := runtime.New()
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			pkgPath, err := d.pkg.GetPkgPath(rootDir, rt)
			if err != nil {
				t.Fatal(err)
			}
			if pkgPath != d.exp {
				t.Fatalf("wanted %v, got %v", d.exp, pkgPath)
			}
		})
	}
}
