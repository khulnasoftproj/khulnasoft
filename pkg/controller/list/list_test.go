package list_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config"
	finder "github.com/khulnasoftproj/khulnasoft/v2/pkg/config-finder"
	reader "github.com/khulnasoftproj/khulnasoft/v2/pkg/config-reader"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/controller/list"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/cosign"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/download"
	registry "github.com/khulnasoftproj/khulnasoft/v2/pkg/install-registry"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/runtime"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/slsa"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/testutil"
	"github.com/sirupsen/logrus"
)

func TestController_List(t *testing.T) {
	t.Parallel()
	data := []struct {
		name              string
		files             map[string]string
		param             *config.Param
		registryInstaller registry.Installer
		isErr             bool
	}{
		{
			name: "normal",
			param: &config.Param{
				PWD:            "/home/foo/workspace",
				ConfigFilePath: "khulnasoft.yaml",
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
				"/home/foo/workspace/registry.yaml": `packages:
- type: github_content
  repo_owner: khulnasoftproj
  repo_name: khulnasoft-installer
  path: khulnasoft-installer
`,
			},
		},
	}
	logE := logrus.NewEntry(logrus.New())
	ctx := context.Background()
	downloader := download.NewGitHubContentFileDownloader(nil, download.NewHTTPDownloader(http.DefaultClient))
	rt := &runtime.Runtime{}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			fs, err := testutil.NewFs(d.files)
			if err != nil {
				t.Fatal(err)
			}
			ctrl := list.NewController(finder.NewConfigFinder(fs), reader.New(fs, d.param), registry.New(d.param, downloader, fs, rt, &cosign.MockVerifier{}, &slsa.MockVerifier{}), fs)
			if err := ctrl.List(ctx, d.param, logE); err != nil {
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
