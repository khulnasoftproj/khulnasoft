package initcmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/khulnasoftproj/khulnasoft/v2/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

const configTemplate = `---
# khulnasoft - Declarative CLI Version Manager
# https://khulnasoftproj.github.io/
# checksum:
#   enabled: true
#   require_checksum: true
#   supported_envs:
#   - all
registries:
- type: standard
  ref: %%STANDARD_REGISTRY_VERSION%% # renovate: depName=khulnasoftproj/khulnasoft-registry
packages:
`

type Controller struct {
	github RepositoriesService
	fs     afero.Fs
}

func New(gh RepositoriesService, fs afero.Fs) *Controller {
	return &Controller{
		github: gh,
		fs:     fs,
	}
}

func (ctrl *Controller) Init(ctx context.Context, cfgFilePath string, logE *logrus.Entry) error {
	if cfgFilePath == "" {
		cfgFilePath = "khulnasoft.yaml"
	}
	if _, err := ctrl.fs.Stat(cfgFilePath); err == nil {
		// configuration file already exists, then do nothing.
		logE.WithFields(logrus.Fields{
			"configuration_file_path": cfgFilePath,
		}).Info("configuration file already exists")
		return nil
	}

	registryVersion := "v4.40.0" // renovate: depName=khulnasoftproj/khulnasoft-registry
	release, _, err := ctrl.github.GetLatestRelease(ctx, "khulnasoftproj", "khulnasoft-registry")
	if err != nil {
		logerr.WithError(logE, err).WithFields(logrus.Fields{
			"repo_owner": "khulnasoftproj",
			"repo_name":  "khulnasoft-registry",
		}).Warn("get the latest release")
	} else {
		if release == nil {
			logE.WithFields(logrus.Fields{
				"repo_owner": "khulnasoftproj",
				"repo_name":  "khulnasoft-registry",
			}).Warn("failed to get the latest release")
		} else {
			registryVersion = release.GetTagName()
		}
	}
	cfgStr := strings.Replace(configTemplate, "%%STANDARD_REGISTRY_VERSION%%", registryVersion, 1)
	if err := afero.WriteFile(ctrl.fs, cfgFilePath, []byte(cfgStr), util.FilePermission); err != nil {
		return fmt.Errorf("write a configuration file: %w", logerr.WithFields(err, logrus.Fields{
			"configuration_file_path": cfgFilePath,
		}))
	}
	return nil
}
