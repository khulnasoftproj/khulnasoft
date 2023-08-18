package cli

import (
	"fmt"

	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/controller"
	"github.com/urfave/cli/v2"
)

func (runner *Runner) newInitCommand() *cli.Command {
	return &cli.Command{
		Name:      "init",
		Usage:     "Create a configuration file if it doesn't exist",
		ArgsUsage: `[<created file path. The default value is "khulnasoft.yaml">]`,
		Description: `Create a configuration file if it doesn't exist
e.g.
$ khulnasoft init # create "khulnasoft.yaml"
$ khulnasoft init foo.yaml # create foo.yaml`,
		Action: runner.initAction,
	}
}

func (runner *Runner) initAction(c *cli.Context) error {
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
	if err := runner.setParam(c, "init", param); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeInitCommandController(c.Context, param)
	return ctrl.Init(c.Context, c.Args().First(), runner.LogE) //nolint:wrapcheck
}
