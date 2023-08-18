package updatekhulnasoft

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/runtime"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type Controller struct {
	rootDir   string
	fs        afero.Fs
	runtime   *runtime.Runtime
	github    RepositoriesService
	installer KhulnasoftInstaller
}

func New(param *config.Param, fs afero.Fs, rt *runtime.Runtime, gh RepositoriesService, installer KhulnasoftInstaller) *Controller {
	return &Controller{
		rootDir:   param.RootDir,
		fs:        fs,
		runtime:   rt,
		github:    gh,
		installer: installer,
	}
}

func (ctrl *Controller) UpdateKhulnasoft(ctx context.Context, logE *logrus.Entry, param *config.Param) error {
	rootBin := filepath.Join(ctrl.rootDir, "bin")
	if err := util.MkdirAll(ctrl.fs, rootBin); err != nil {
		return fmt.Errorf("create the directory: %w", err)
	}

	version, err := ctrl.getVersion(ctx, param)
	if err != nil {
		return err
	}

	logE = logE.WithField("new_version", version)

	if err := ctrl.installer.InstallKhulnasoft(ctx, logE, version); err != nil {
		return fmt.Errorf("download khulnasoft: %w", logerr.WithFields(err, logrus.Fields{
			"new_version": version,
		}))
	}
	return nil
}

func (ctrl *Controller) getVersion(ctx context.Context, param *config.Param) (string, error) {
	switch len(param.Args) {
	case 0:
		release, _, err := ctrl.github.GetLatestRelease(ctx, "khulnasoftproj", "khulnasoft")
		if err != nil {
			return "", fmt.Errorf("get the latest version of khulnasoft: %w", err)
		}
		return release.GetTagName(), nil
	case 1:
		return param.Args[0], nil
	default:
		return "", errors.New("too many arguments")
	}
}
