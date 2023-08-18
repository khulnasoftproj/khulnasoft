package checksum_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/checksum"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/testutil"
	"github.com/spf13/afero"
)

func TestChecksums_Get(t *testing.T) {
	t.Parallel()
	data := []struct {
		name string
		m    map[string]*checksum.Checksum
		key  string
		exp  *checksum.Checksum
	}{
		{
			name: "key not found",
			key:  "foo",
			exp:  nil,
		},
		{
			name: "key is found",
			key:  "foo",
			m: map[string]*checksum.Checksum{
				"foo": {
					ID:        "foo",
					Checksum:  "bar",
					Algorithm: "sha256",
				},
			},
			exp: &checksum.Checksum{
				ID:        "foo",
				Checksum:  "BAR",
				Algorithm: "sha256",
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			checksums := checksum.New()
			for k, v := range d.m {
				checksums.Set(k, v)
			}
			v := checksums.Get(d.key)
			if diff := cmp.Diff(v, d.exp); diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}

func TestChecksums_ReadFile(t *testing.T) {
	t.Parallel()
	data := []struct {
		name  string
		m     map[string]string
		p     string
		isErr bool
	}{
		{
			name: "file not found",
			p:    ".khulnasoft-checksums.json",
		},
		{
			name: "file is found",
			m: map[string]string{
				".khulnasoft-checksums.json": `{
  "github_release/github.com/cli/cli/v2.10.1/gh_2.10.1_macOS_amd64.tar.gz": "xxx"
}`,
			},
			p: ".khulnasoft-checksums.json",
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			fs, err := testutil.NewFs(d.m)
			if err != nil {
				t.Fatal(err)
			}
			checksums := checksum.New()
			if err := checksums.ReadFile(fs, d.p); err != nil {
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

func TestChecksums_UpdateFile(t *testing.T) {
	t.Parallel()
	data := []struct {
		name  string
		m     []*checksum.Checksum
		p     string
		isErr bool
	}{
		{
			name: "normal",
			m: []*checksum.Checksum{
				{
					ID:        "github_release/github.com/cli/cli/v2.10.1/gh_2.10.1_macOS_amd64.tar.gz",
					Checksum:  "xxx",
					Algorithm: "sha256",
				},
			},
			p: ".khulnasoft-checksums.json",
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			fs := afero.NewMemMapFs()
			checksums := checksum.New()
			for _, v := range d.m {
				checksums.Set(v.ID, v)
			}
			if err := checksums.UpdateFile(fs, d.p); err != nil {
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

func TestGetChecksumFilePathFromConfigFilePath(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name        string
		cfgFilePath string
		exp         string
		files       map[string]string
	}{
		{
			name:        "new",
			cfgFilePath: "khulnasoft.yaml",
			exp:         "khulnasoft-checksums.json",
		},
		{
			name:        "khulnasoft-checksums.json > .khulnasoft-checksums.json",
			cfgFilePath: "khulnasoft.yaml",
			exp:         "khulnasoft-checksums.json",
			files: map[string]string{
				"khulnasoft-checksums.json":  "",
				".khulnasoft-checksums.json": "",
			},
		},
		{
			name:        ".khulnasoft-checksums.json",
			cfgFilePath: "khulnasoft.yaml",
			exp:         ".khulnasoft-checksums.json",
			files: map[string]string{
				".khulnasoft-checksums.json": "",
			},
		},
		{
			name:        "new absolute",
			cfgFilePath: "/home/foo/khulnasoft.yaml",
			exp:         "/home/foo/khulnasoft-checksums.json",
		},
		{
			name:        "absolute khulnasoft-checksums.json > .khulnasoft-checksums.json",
			cfgFilePath: "/home/foo/khulnasoft.yaml",
			exp:         "/home/foo/khulnasoft-checksums.json",
			files: map[string]string{
				"/home/foo/.khulnasoft-checksums.json": "",
				"/home/foo/khulnasoft-checksums.json":  "",
			},
		},
		{
			name:        "absolute .khulnasoft-checksums.json",
			cfgFilePath: "/home/foo/khulnasoft.yaml",
			exp:         "/home/foo/.khulnasoft-checksums.json",
			files: map[string]string{
				"/home/foo/.khulnasoft-checksums.json": "",
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
			p, err := checksum.GetChecksumFilePathFromConfigFilePath(fs, d.cfgFilePath)
			if err != nil {
				t.Fatal(err)
			}
			if p != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, p)
			}
		})
	}
}
