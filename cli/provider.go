package cli

import (
	"fmt"
	"github.com/urfave/cli/v2"
)

var providerCmds = &cli.Command{
	Name:  "provider",
	Usage: "Manage provider",
	Subcommands: []*cli.Command{
		ProviderList,
	},
}

var ProviderList = &cli.Command{
	Name:  "list",
	Usage: "List providers",
	Action: func(cctx *cli.Context) error {
		api, closer, err := GetManagerAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()

		ctx := ReqContext(cctx)

		providers, err := api.GetProviderList(ctx)
		if err != nil {
			return err
		}

		for _, provider := range providers {
			fmt.Println(provider)
		}

		return nil
	},
}
