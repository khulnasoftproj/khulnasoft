package installpackage_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/khulnasoftproj/khulnasoft/v2/pkg/checksum"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config"
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
)

func Test_installer_InstallProxy(t *testing.T) {
	t.Parallel()
	data := []struct {
		name     string
		files    map[string]string
		param    *config.Param
		rt       *runtime.Runtime
		executor installpackage.Executor
		links    map[string]string
		isErr    bool
	}{
		{
			name: "file already exists",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			param: &config.Param{
				RootDir:        "/home/foo/.local/share/khulnasoftproj-khulnasoft",
				PWD:            "/home/foo/workspace",
				ConfigFilePath: "khulnasoft.yaml",
				MaxParallelism: 5,
			},
			files: map[string]string{
				fmt.Sprintf("/home/foo/.local/share/khulnasoftproj-khulnasoft/internal/pkgs/github_release/github.com/khulnasoftproj/khulnasoft-proxy/%s/khulnasoft-proxy_linux_amd64.tar.gz/khulnasoft-proxy", installpackage.ProxyVersion): "",
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
			downloader := download.NewDownloader(nil, download.NewHTTPDownloader(http.DefaultClient))
			ctrl := installpackage.New(d.param, downloader, d.rt, fs, linker, nil, &checksum.Calculator{}, unarchive.New(d.executor, fs), &policy.Checker{}, &cosign.MockVerifier{}, &slsa.MockVerifier{}, &installpackage.MockGoInstallInstaller{}, &installpackage.MockGoBuildInstaller{}, &installpackage.MockCargoPackageInstaller{})
			if err := ctrl.InstallProxy(ctx, logE); err != nil {
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
