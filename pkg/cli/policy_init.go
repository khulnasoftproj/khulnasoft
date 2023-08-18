package cli

import (
	"github.com/urfave/cli/v2"
)

func (runner *Runner) newPolicyInitCommand() *cli.Command {
	return &cli.Command{
		Name:      "init",
		Usage:     "Create a policy file if it doesn't exist",
		ArgsUsage: `[<created file path. The default value is "khulnasoft-policy.yaml">]`,
		Description: `Create a policy file if it doesn't exist
e.g.
$ khulnasoft policy init # create "khulnasoft-policy.yaml"
$ khulnasoft policy init foo.yaml # create foo.yaml`,
		Action: runner.initPolicyAction,
	}
}
