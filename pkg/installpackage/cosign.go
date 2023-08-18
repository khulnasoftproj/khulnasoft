package installpackage

import (
	"context"
	"fmt"
	"sync"

	"github.com/khulnasoftproj/khulnasoft/v2/pkg/checksum"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config/khulnasoft"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config/registry"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/cosign"
	"github.com/sirupsen/logrus"
)

type Cosign struct {
	installer *InstallerImpl
	mutex     *sync.Mutex
}

func (cos *Cosign) installCosign(ctx context.Context, logE *logrus.Entry, version string) error {
	cos.mutex.Lock()
	defer cos.mutex.Unlock()
	assetTemplate := `cosign-{{.OS}}-{{.Arch}}`
	pkg := &config.Package{
		Package: &khulnasoft.Package{
			Name:    "sigstore/cosign",
			Version: version,
		},
		PackageInfo: &registry.PackageInfo{
			Type:      "github_release",
			RepoOwner: "sigstore",
			RepoName:  "cosign",
			Asset:     &assetTemplate,
			SupportedEnvs: []string{
				"darwin",
				"linux",
				"amd64",
			},
		},
	}

	chksum := cosign.Checksums()[cos.installer.runtime.Env()]

	pkgInfo, err := pkg.PackageInfo.Override(logE, pkg.Package.Version, cos.installer.runtime)
	if err != nil {
		return fmt.Errorf("evaluate version constraints: %w", err)
	}
	supported, err := pkgInfo.CheckSupported(cos.installer.runtime, cos.installer.runtime.Env())
	if err != nil {
		return fmt.Errorf("check if cosign is supported: %w", err)
	}
	if !supported {
		logE.Debug("the package isn't supported on this environment")
		return nil
	}

	pkg.PackageInfo = pkgInfo

	if err := cos.installer.InstallPackage(ctx, logE, &ParamInstallPackage{
		Checksums: checksum.New(), // Check cosign's checksum but not update khulnasoft-checksums.json
		Pkg:       pkg,
		Checksum: &checksum.Checksum{
			Algorithm: "sha256",
			Checksum:  chksum,
		},
		DisablePolicy: true,
	}); err != nil {
		return err
	}

	return nil
}
