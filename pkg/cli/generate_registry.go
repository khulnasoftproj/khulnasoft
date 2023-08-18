package cli

import (
	"fmt"
	"net/http"
	"os"

	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/controller"
	"github.com/urfave/cli/v2"
)

const generateRegistryDescription = `Generate a template of Registry package configuration.

Note that you probably fix the generate code manually.
The generate code is not perfect and may include the wrong configuration.
It is just a template.

e.g.

$ khulnasoft gr cli/cli # Outputs the configuration.
packages:
  - type: github_release
    repo_owner: cli
    repo_name: cli
    asset: gh_{{trimV .Version}}_{{.OS}}_{{.Arch}}.{{.Format}}
    format: tar.gz
    description: GitHub’s official command line tool
    replacements:
      darwin: macOS
    overrides:
      - goos: windows
        format: zip
    supported_envs:
      - darwin
      - linux
      - amd64
    rosetta2: true

By default, khulnasoft gets the information from the latest GitHub Releases.
You can specify a specific package version.

e.g.

$ khulnasoft gr cli/cli@v2.0.0

By default, khulnasoft gr doesn't generate version_overrides.
If --deep is set, khulnasoft generates version_overrides.

e.g.

$ khulnasoft gr --deep sulaiman-coder/tfcmt

Note that if --deep is set, GitHub API is called per GitHub Release.
This may cause GitHub API rate limiting.

If --out-testdata is set, khulnasoft inserts testdata into the specified file.

e.g.

$ khulnasoft gr --out-testdata testdata.yaml sulaiman-coder/tfcmt
`

func (runner *Runner) newGenerateRegistryCommand() *cli.Command {
	return &cli.Command{
		Name:        "generate-registry",
		Aliases:     []string{"gr"},
		Usage:       "Generate a registry's package configuration",
		ArgsUsage:   `<package name>`,
		Description: generateRegistryDescription,
		Action:      runner.generateRegistryAction,
		// TODO support "i" option
		Flags: []cli.Flag{
			// 	&cli.StringFlag{
			// 		Name:  "i",
			// 		Usage: "Insert a registry to configuration file",
			// 	},
			&cli.StringFlag{
				Name:  "out-testdata",
				Usage: "A file path where the testdata is outputted",
			},
			&cli.BoolFlag{
				Name:  "deep",
				Usage: "Resolve version_overrides",
			},
		},
	}
}

func (runner *Runner) generateRegistryAction(c *cli.Context) error {
	tracer, err := startTrace(c.String("trace"))
	if err != nil {
		return err
	}
	defer tracer.Stop()

	cpuProfiler, err := startCPUProfile(c.String("cpu-profile"))
	if err != nil {
		return err
	}
	defer cpuProfiler.Stop()

	param := &config.Param{}
	if err := runner.setParam(c, "generate-registry", param); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeGenerateRegistryCommandController(c.Context, param, http.DefaultClient, os.Stdout)
	return ctrl.GenerateRegistry(c.Context, param, runner.LogE, c.Args().Slice()...) //nolint:wrapcheck
}
