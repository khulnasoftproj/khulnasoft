package config_test

import (
	"testing"

	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
)

func TestGetRootDir(t *testing.T) {
	t.Parallel()
	data := []struct {
		name string
		env  map[string]string
		exp  string
	}{
		{
			name: "KHULNASOFT_ROOT_DIR",
			env: map[string]string{
				"KHULNASOFT_ROOT_DIR": "/home/foo/.khulnasoft",
			},
			exp: "/home/foo/.khulnasoft",
		},
		{
			name: "XDG_DATA_HOME",
			env: map[string]string{
				"XDG_DATA_HOME": "/home/foo/.xdg",
			},
			exp: "/home/foo/.xdg/khulnasoftproj-khulnasoft",
		},
		{
			name: "HOME",
			env: map[string]string{
				"HOME": "/home/foo",
			},
			exp: "/home/foo/.local/share/khulnasoftproj-khulnasoft",
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			rootDir := config.GetRootDir(osenv.NewMock(d.env))
			if rootDir != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, rootDir)
			}
		})
	}
}
