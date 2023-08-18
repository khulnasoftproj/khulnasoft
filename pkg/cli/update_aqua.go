package cli

import (
	"fmt"
	"net/http"

	"github.com/khulnasoftproj/khulnasoft/v2/pkg/config"
	"github.com/khulnasoftproj/khulnasoft/v2/pkg/controller"
	"github.com/urfave/cli/v2"
)

func (runner *Runner) newUpdateKhulnasoftCommand() *cli.Command {
	return &cli.Command{
		Name:  "update-khulnasoft",
		Usage: "Update khulnasoft",
		Description: `Update khulnasoft.

e.g.
$ khulnasoft update-khulnasoft [version]

khulnasoft is installed in $KHULNASOFT_ROOT_DIR/bin.
By default the latest version of khulnasoft is installed, but you can specify the version with argument.

e.g.
$ khulnasoft update-khulnasoft # Install the latest version
$ khulnasoft update-khulnasoft v1.20.0 # Install v1.20.0
`,
		Action: runner.updaetKhulnasoftAction,
	}
}

func (runner *Runner) updaetKhulnasoftAction(c *cli.Context) error {
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
	if err := runner.setParam(c, "update-khulnasoft", param); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeUpdateKhulnasoftCommandController(c.Context, param, http.DefaultClient, runner.Runtime)
	return ctrl.UpdateKhulnasoft(c.Context, runner.LogE, param) //nolint:wrapcheck
}
