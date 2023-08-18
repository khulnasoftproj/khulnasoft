package cosign_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config/registry"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/cosign"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/download"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/exec"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/runtime"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/template"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func TestVerifier_Verify(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name             string
		isErr            bool
		executor         cosign.Executor
		fs               afero.Fs
		downloader       download.ClientAPI
		cosignExePath    string
		rt               *runtime.Runtime
		param            *config.Param
		file             *download.File
		cos              *registry.Cosign
		art              *template.Artifact
		verifiedFilePath string
	}{
		{
			name:     "normal",
			executor: &exec.Mock{},
			fs:       afero.NewMemMapFs(),
			downloader: &download.Mock{
				RC: io.NopCloser(strings.NewReader("hello")),
			},
			cosignExePath: "/home/foo/.local/share/khulnasoftproj-khulnasoft/pkgs/github_release/github.com/sigstore/cosign/v1.13.1/cosign-darwin-arm64/cosign-darwin-arm64",
			rt: &runtime.Runtime{
				GOOS:   "darwin",
				GOARCH: "arm64",
			},
			file: &download.File{
				Type:      "github_content",
				RepoOwner: "khulnasoftproj",
				RepoName:  "khulnasoft-installer",
				Version:   "v1.1.3",
				Path:      "khulnasoft-installer",
			},
			param: &config.Param{
				RootDir: "/home/foo/.local/share/khulnasoftproj-khulnasoft",
			},
			cos: &registry.Cosign{
				CosignExperimental: true,
				Opts: []string{
					"--signature",
					"https://github.com/khulnasoftproj/khulnasoft-installer/releases/download/{{.Version}}/khulnasoft-installer.sig",
					"--certificate",
					"https://github.com/khulnasoftproj/khulnasoft-installer/releases/download/{{.Version}}/khulnasoft-installer.pem",
				},
			},
			art: &template.Artifact{
				Version: "v1.1.3",
				OS:      "darwin",
				Arch:    "arm64",
				Format:  "raw",
				Asset:   "khulnasoft-installer",
			},
		},
		{
			name:     "signature, key, certificate",
			executor: &exec.Mock{},
			fs:       afero.NewMemMapFs(),
			downloader: &download.Mock{
				RC: io.NopCloser(strings.NewReader("hello")),
			},
			cosignExePath: "/home/foo/.local/share/khulnasoftproj-khulnasoft/pkgs/github_release/github.com/sigstore/cosign/v1.13.1/cosign-darwin-arm64/cosign-darwin-arm64",
			rt: &runtime.Runtime{
				GOOS:   "darwin",
				GOARCH: "arm64",
			},
			file: &download.File{
				Type:      "github_content",
				RepoOwner: "khulnasoftproj",
				RepoName:  "khulnasoft-installer",
				Version:   "v1.1.3",
				Path:      "khulnasoft-installer",
			},
			param: &config.Param{
				RootDir: "/home/foo/.local/share/khulnasoftproj-khulnasoft",
			},
			cos: &registry.Cosign{
				CosignExperimental: true,
				Signature: &registry.DownloadedFile{
					Type:  "github_release",
					Asset: util.StrP("khulnasoft-installer.sig"),
				},
				Certificate: &registry.DownloadedFile{
					Type:  "github_release",
					Asset: util.StrP("khulnasoft-installer.pem"),
				},
				Key: &registry.DownloadedFile{
					Type:  "github_release",
					Asset: util.StrP("khulnasoft-installer.key"),
				},
			},
			art: &template.Artifact{
				Version: "v1.1.3",
				OS:      "darwin",
				Arch:    "arm64",
				Format:  "raw",
				Asset:   "khulnasoft-installer",
			},
		},
	}
	ctx := context.Background()
	logE := logrus.NewEntry(logrus.New())
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			verifier := cosign.NewVerifier(d.executor, d.fs, d.downloader, d.param)
			if err := verifier.Verify(ctx, logE, d.rt, d.file, d.cos, d.art, d.verifiedFilePath); err != nil {
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
