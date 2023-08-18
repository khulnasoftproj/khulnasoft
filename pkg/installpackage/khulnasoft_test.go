package installpackage_test

import (
	"context"
	"io"
	"strings"
	"testing"

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

func Test_installer_InstallKhulnasoft(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name               string
		files              map[string]string
		param              *config.Param
		rt                 *runtime.Runtime
		checksumDownloader download.ChecksumDownloader
		checksumCalculator installpackage.ChecksumCalculator
		version            string
		isTest             bool
		isErr              bool
	}{
		{
			name: "file already exists",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			param: &config.Param{
				RootDir: "/home/foo/.local/share/khulnasoftproj-khulnasoft",
			},
			files: map[string]string{
				"/home/foo/.local/share/khulnasoftproj-khulnasoft/internal/pkgs/github_release/github.com/khulnasoftproj/khulnasoft/v1.6.1/khulnasoft_linux_amd64.tar.gz/khulnasoft": "xxx",
			},
			version: "v1.6.1",
			checksumCalculator: &installpackage.MockChecksumCalculator{
				Checksum: "c6f3b1f37d9bf4f73e6c6dcf1bd4bb59b48447ad46d4b72e587d15f66a96ab5a",
			},
			checksumDownloader: &download.MockChecksumDownloader{
				Body: `31adc2cfc3aab8e66803f6769016fe6953a22f88de403211abac83c04a542d46  khulnasoft_darwin_arm64.tar.gz
6e53f151abf10730bdfd4a52b99019ffa5f58d8ad076802affb3935dd82aba96  khulnasoft_darwin_amd64.tar.gz
c6f3b1f37d9bf4f73e6c6dcf1bd4bb59b48447ad46d4b72e587d15f66a96ab5a  khulnasoft_linux_amd64.tar.gz
e922723678f493216c2398f3f23fb027c9a98808b49f6fce401ef82ee2c22b03  khulnasoft_linux_arm64.tar.gz`,
				Code: 200,
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
			ctrl := installpackage.New(d.param, &download.Mock{
				RC: io.NopCloser(strings.NewReader("xxx")),
			}, d.rt, fs, domain.NewMockLinker(fs), d.checksumDownloader, d.checksumCalculator, &unarchive.MockUnarchiver{}, &policy.Checker{}, &cosign.MockVerifier{}, &slsa.MockVerifier{}, &installpackage.MockGoInstallInstaller{}, &installpackage.MockGoBuildInstaller{}, &installpackage.MockCargoPackageInstaller{})
			if err := ctrl.InstallKhulnasoft(ctx, logE, d.version); err != nil {
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
