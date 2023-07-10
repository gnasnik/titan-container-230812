package cli

import "github.com/urfave/cli/v2"

var providersCmds = &cli.Command{
	Name:        "providers",
	Usage:       "List providers",
	Subcommands: []*cli.Command{},
}
