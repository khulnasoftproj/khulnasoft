//go:build !windows
// +build !windows

package config

import (
	"path/filepath"

	"github.com/suzuki-shunsuke/go-osenv/osenv"
)

func GetRootDir(osEnv osenv.OSEnv) string {
	if rootDir := osEnv.Getenv("KHULNASOFT_ROOT_DIR"); rootDir != "" {
		return rootDir
	}
	xdgDataHome := osEnv.Getenv("XDG_DATA_HOME")
	if xdgDataHome == "" {
		xdgDataHome = filepath.Join(osEnv.Getenv("HOME"), ".local", "share")
	}
	return filepath.Join(xdgDataHome, "khulnasoftproj-khulnasoft")
}
