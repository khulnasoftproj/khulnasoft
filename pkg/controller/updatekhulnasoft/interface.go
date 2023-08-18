package updatekhulnasoft

import (
	"context"

	"github.com/khulnasoftproj/khulnasoft/v2/pkg/github"
	"github.com/sirupsen/logrus"
)

type KhulnasoftInstaller interface {
	InstallKhulnasoft(ctx context.Context, logE *logrus.Entry, version string) error
}

type RepositoriesService interface {
	GetLatestRelease(ctx context.Context, repoOwner, repoName string) (*github.RepositoryRelease, *github.Response, error)
}

type ConfigFinder interface {
	Finds(wd, configFilePath string) []string
}
