package cli

import (
	"github.com/gnasnik/titan-container/api/types"
	"github.com/gnasnik/titan-container/lib/tablewriter"
	"github.com/urfave/cli/v2"
	"os"
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

		tw := tablewriter.New(
			tablewriter.Col("ID"),
			tablewriter.Col("IP"),
			tablewriter.Col("State"),
			tablewriter.Col("HostURI"),
			tablewriter.Col("CreatedTime"),
		)

		providers, err := api.GetProviderList(ctx)
		if err != nil {
			return err
		}

		for _, provider := range providers {
			m := map[string]interface{}{
				"ID":         provider.ID,
				"IP":         provider.IP,
				"State":      types.ProviderStateString(provider.State),
				"HostURI":    provider.HostURI,
				"CreateTime": provider.CreatedAt,
			}
			tw.Write(m)
		}

		tw.Flush(os.Stdout)
		return nil
	},
}
