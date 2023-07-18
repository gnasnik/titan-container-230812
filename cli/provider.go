package cli

import (
	"github.com/docker/go-units"
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
			tablewriter.Col("CPU"),
			tablewriter.Col("CPUAvail"),
			tablewriter.Col("Memory"),
			tablewriter.Col("MemoryAvail"),
			tablewriter.Col("Storage"),
			tablewriter.Col("StorageAvail"),
			tablewriter.Col("CreatedTime"),
		)

		providers, err := api.GetProviderList(ctx)
		if err != nil {
			return err
		}

		for _, provider := range providers {
			resource, err := api.GetStatistics(ctx, provider.ID)
			if err != nil {
				continue
			}

			m := map[string]interface{}{
				"ID":           provider.ID,
				"IP":           provider.IP,
				"State":        types.ProviderStateString(provider.State),
				"HostURI":      provider.HostURI,
				"CPU":          resource.CPUCores.MaxCPUCores,
				"CPUAvail":     resource.CPUCores.Available,
				"Memory":       units.BytesSize(float64(resource.Memory.MaxMemory)),
				"MemoryAvail":  units.BytesSize(float64(resource.Memory.Available)),
				"Storage":      units.BytesSize(float64(resource.Storage.MaxStorage)),
				"StorageAvail": units.BytesSize(float64(resource.Storage.Available)),
				"CreateTime":   provider.CreatedAt.Format(defaultDateTimeLayout),
			}
			tw.Write(m)
		}

		tw.Flush(os.Stdout)
		return nil
	},
}
