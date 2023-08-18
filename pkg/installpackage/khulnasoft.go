package installpackage

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/khulnasoftproj/khulnasoft/v2/pkg/checksum"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config/khulnasoft"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config/registry"
	"github.com/sirupsen/logrus"
)

func (inst *InstallerImpl) InstallKhulnasoft(ctx context.Context, logE *logrus.Entry, version string) error { //nolint:funlen
	assetTemplate := `khulnasoft_{{.OS}}_{{.Arch}}.tar.gz`
	provTemplate := "multiple.intoto.jsonl"
	disabled := false
	pkg := &config.Package{
		Package: &khulnasoft.Package{
			Name:    "khulnasoftproj/khulnasoft",
			Version: version,
		},
		PackageInfo: &registry.PackageInfo{
			Type:      "github_release",
			RepoOwner: "khulnasoftproj",
			RepoName:  "khulnasoft",
			Asset:     &assetTemplate,
			Files: []*registry.File{
				{
					Name: "khulnasoft",
				},
			},
			SLSAProvenance: &registry.SLSAProvenance{
				Type:  "github_release",
				Asset: &provTemplate,
			},
			// Checksum: &registry.Checksum{
			// 	Type:       "github_release",
			// 	Asset:      "khulnasoft_{{trimV .Version}}_checksums.txt",
			// 	FileFormat: "regexp",
			// 	Algorithm:  "sha256",
			// 	Pattern: &registry.ChecksumPattern{
			// 		Checksum: `^(\b[A-Fa-f0-9]{64}\b)`,
			// 		File:     `^\b[A-Fa-f0-9]{64}\b\s+(\S+)$`,
			// 	},
			// 	Cosign: &registry.Cosign{
			// 		CosignExperimental: true,
			// 		Opts: []string{
			// 			"--signature",
			// 			"https://github.com/khulnasoftproj/khulnasoft/releases/download/{{.Version}}/khulnasoft_{{trimV .Version}}_checksums.txt.sig",
			// 			"--certificate",
			// 			"https://github.com/khulnasoftproj/khulnasoft/releases/download/{{.Version}}/khulnasoft_{{trimV .Version}}_checksums.txt.pem",
			// 		},
			// 	},
			// },
			VersionConstraints: `semver(">= 1.26.0")`,
			VersionOverrides: []*registry.VersionOverride{
				{
					VersionConstraints: "true",
					SLSAProvenance: &registry.SLSAProvenance{
						Enabled: &disabled,
					},
					Checksum: &registry.Checksum{
						Type:       "github_release",
						Asset:      "khulnasoft_{{trimV .Version}}_checksums.txt",
						FileFormat: "regexp",
						Algorithm:  "sha256",
						Pattern: &registry.ChecksumPattern{
							Checksum: `^(\b[A-Fa-f0-9]{64}\b)`,
							File:     `^\b[A-Fa-f0-9]{64}\b\s+(\S+)$`,
						},
						Cosign: &registry.Cosign{
							Opts: []string{},
						},
					},
				},
			},
		},
	}

	pkgInfo, err := pkg.PackageInfo.Override(logE, pkg.Package.Version, inst.runtime)
	if err != nil {
		return fmt.Errorf("evaluate version constraints: %w", err)
	}
	pkg.PackageInfo = pkgInfo

	if err := inst.InstallPackage(ctx, logE, &ParamInstallPackage{
		Checksums:     checksum.New(), // Check khulnasoft's checksum but not update khulnasoft-checksums.json
		Pkg:           pkg,
		DisablePolicy: true,
	}); err != nil {
		return err
	}

	logE = logE.WithFields(logrus.Fields{
		"package_name":    pkg.Package.Name,
		"package_version": pkg.Package.Version,
	})

	exePath, err := pkg.GetExePath(inst.rootDir, &registry.File{
		Name: "khulnasoft",
	}, inst.runtime)
	if err != nil {
		return fmt.Errorf("get the executable file path: %w", err)
	}

	if inst.runtime.GOOS == "windows" {
		return inst.Copy(filepath.Join(inst.rootDir, "bin", "khulnasoft.exe"), exePath)
	}

	// create a symbolic link
	a, err := filepath.Rel(filepath.Join(inst.rootDir, "bin"), exePath)
	if err != nil {
		return fmt.Errorf("get a relative path: %w", err)
	}

	return inst.createLink(filepath.Join(inst.rootDir, "bin", "khulnasoft"), a, logE)
}
