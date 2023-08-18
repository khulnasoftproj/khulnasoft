package policy_test

import (
	"testing"

	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/policy"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func TestValidator_Allow(t *testing.T) { //nolint:dupl
	t.Parallel()
	data := []struct {
		name           string
		rootDir        string
		configFilePath string
		files          map[string]string
		isErr          bool
	}{
		{
			name:           "normal",
			rootDir:        "/home/foo/.local/share/khulnasoftproj-khulnasoft",
			configFilePath: "/home/foo/workspace/khulnasoft-policy.yaml",
			files: map[string]string{
				"/home/foo/workspace/khulnasoft-policy.yaml": "",
			},
		},
		{
			name:           "warn file exists",
			rootDir:        "/home/foo/.local/share/khulnasoftproj-khulnasoft",
			configFilePath: "/home/foo/workspace/khulnasoft-policy.yaml",
			files: map[string]string{
				"/home/foo/workspace/khulnasoft-policy.yaml":                                                                 "",
				"/home/foo/.local/share/khulnasoftproj-khulnasoft/policy-warnings/home/foo/workspace/khulnasoft-policy.yaml": "",
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			fs := afero.NewMemMapFs()
			for name, body := range d.files {
				if err := afero.WriteFile(fs, name, []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			validator := policy.NewValidator(&config.Param{
				RootDir: d.rootDir,
			}, fs)
			if err := validator.Allow(d.configFilePath); err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returend")
			}
		})
	}
}

func TestValidator_Deny(t *testing.T) { //nolint:dupl
	t.Parallel()
	data := []struct {
		name           string
		rootDir        string
		configFilePath string
		files          map[string]string
		isErr          bool
	}{
		{
			name:           "normal",
			rootDir:        "/home/foo/.local/share/khulnasoftproj-khulnasoft",
			configFilePath: "/home/foo/workspace/khulnasoft-policy.yaml",
			files: map[string]string{
				"/home/foo/workspace/khulnasoft-policy.yaml": "",
			},
		},
		{
			name:           "remove allowed policy file",
			configFilePath: "/home/foo/workspace/khulnasoft-policy.yaml",
			rootDir:        "/home/foo/.local/share/khulnasoftproj-khulnasoft",
			files: map[string]string{
				"/home/foo/workspace/khulnasoft-policy.yaml":                                                          "",
				"/home/foo/.local/share/khulnasoftproj-khulnasoft/policies/home/foo/workspace/khulnasoft-policy.yaml": "",
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			fs := afero.NewMemMapFs()
			for name, body := range d.files {
				if err := afero.WriteFile(fs, name, []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			validator := policy.NewValidator(&config.Param{
				RootDir: d.rootDir,
			}, fs)
			if err := validator.Deny(d.configFilePath); err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returend")
			}
		})
	}
}

func TestValidator_Warn(t *testing.T) {
	t.Parallel()
	data := []struct {
		name           string
		rootDir        string
		configFilePath string
		files          map[string]string
		isErr          bool
	}{
		{
			name:           "normal",
			rootDir:        "/home/foo/.local/share/khulnasoftproj-khulnasoft",
			configFilePath: "/home/foo/workspace/khulnasoft-policy.yaml",
			files: map[string]string{
				"/home/foo/workspace/khulnasoft-policy.yaml": "",
			},
		},
		{
			name:           "warn file exists",
			configFilePath: "/home/foo/workspace/khulnasoft-policy.yaml",
			rootDir:        "/home/foo/.local/share/khulnasoftproj-khulnasoft",
			files: map[string]string{
				"/home/foo/workspace/khulnasoft-policy.yaml":                                                                 "",
				"/home/foo/.local/share/khulnasoftproj-khulnasoft/policy-warnings/home/foo/workspace/khulnasoft-policy.yaml": "",
			},
		},
	}
	logE := logrus.NewEntry(logrus.New())
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			fs := afero.NewMemMapFs()
			for name, body := range d.files {
				if err := afero.WriteFile(fs, name, []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			validator := policy.NewValidator(&config.Param{
				RootDir: d.rootDir,
			}, fs)
			if err := validator.Warn(logE, d.configFilePath, false); err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returend")
			}
		})
	}
}

func TestValidator_Validate(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name           string
		rootDir        string
		configFilePath string
		files          map[string]string
		isErr          bool
	}{
		{
			name:           "normal",
			rootDir:        "/home/foo/.local/share/khulnasoftproj-khulnasoft",
			configFilePath: "/home/foo/workspace/khulnasoft-policy.yaml",
			files: map[string]string{
				"/home/foo/workspace/khulnasoft-policy.yaml":                                                          "",
				"/home/foo/.local/share/khulnasoftproj-khulnasoft/policies/home/foo/workspace/khulnasoft-policy.yaml": "",
			},
		},
		{
			name:           "policy is not found",
			configFilePath: "/home/foo/workspace/khulnasoft-policy.yaml",
			rootDir:        "/home/foo/.local/share/khulnasoftproj-khulnasoft",
			files: map[string]string{
				"/home/foo/workspace/khulnasoft-policy.yaml": "",
			},
			isErr: true,
		},
		{
			name:           "policy is changed",
			configFilePath: "/home/foo/workspace/khulnasoft-policy.yaml",
			rootDir:        "/home/foo/.local/share/khulnasoftproj-khulnasoft",
			files: map[string]string{
				"/home/foo/workspace/khulnasoft-policy.yaml":                                                          "",
				"/home/foo/.local/share/khulnasoftproj-khulnasoft/policies/home/foo/workspace/khulnasoft-policy.yaml": "a",
			},
			isErr: true,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			fs := afero.NewMemMapFs()
			for name, body := range d.files {
				if err := afero.WriteFile(fs, name, []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			validator := policy.NewValidator(&config.Param{
				RootDir: d.rootDir,
			}, fs)
			if err := validator.Validate(d.configFilePath); err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returend")
			}
		})
	}
}
