package cli

import (
	"fmt"
	"os"

	"github.com/docker/go-units"
	"github.com/gnasnik/titan-container/api/types"
	"github.com/gnasnik/titan-container/lib/tablewriter"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var defaultDateTimeLayout = "2006-01-02 15:04:05"

var deploymentCmds = &cli.Command{
	Name:  "deployment",
	Usage: "Manager deployment",
	Subcommands: []*cli.Command{
		CreateDeployment,
		DeploymentList,
		DeleteDeployment,
	},
}

var CreateDeployment = &cli.Command{
	Name:  "create",
	Usage: "create new deployment",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "provider-id",
			Usage:    "the provider id",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "owner",
			Usage: "owner address",
		},
		&cli.StringFlag{
			Name:  "name",
			Usage: "deployment name",
		},
		&cli.StringFlag{
			Name:     "image",
			Usage:    "deployment image",
			Required: true,
		},
		&cli.Float64Flag{
			Name:  "cpu",
			Usage: "cpu cores",
		},
		&cli.Int64Flag{
			Name:  "mem",
			Usage: "memory",
		},
		&cli.Int64Flag{
			Name:  "storage",
			Usage: "storage",
		},
	},
	Action: func(cctx *cli.Context) error {
		api, closer, err := GetManagerAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()

		ctx := ReqContext(cctx)
		providerID := types.ProviderID(cctx.String("provider-id"))

		deployment := &types.Deployment{
			ProviderID: providerID,
			Name:       cctx.String("name"),
			Services: []*types.Service{
				{
					Image: cctx.String("image"),
					ComputeResources: types.ComputeResources{
						CPU:     cctx.Float64("cpu"),
						Memory:  cctx.Int64("mem"),
						Storage: cctx.Int64("storage"),
					},
					Env:       map[string]string{"hello": "test"},
					Arguments: []string{"--test"},
				},
			},
		}

		return api.CreateDeployment(ctx, deployment)
	},
}

var DeploymentList = &cli.Command{
	Name:  "list",
	Usage: "List deployments",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "owner",
			Usage: "owner address",
		},
		&cli.StringFlag{
			Name:  "id",
			Usage: "the deployment id",
		},
		&cli.BoolFlag{
			Name:  "show-all",
			Usage: "show deleted and inactive deployments",
		},
	},
	Action: func(cctx *cli.Context) error {
		api, closer, err := GetManagerAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()

		ctx := ReqContext(cctx)

		tw := tablewriter.New(
			tablewriter.Col("ID"),
			tablewriter.Col("Name"),
			tablewriter.Col("Image"),
			tablewriter.Col("State"),
			tablewriter.Col("CPU"),
			tablewriter.Col("Memory"),
			tablewriter.Col("Storage"),
			tablewriter.Col("CreatedTime"),
			tablewriter.Col("ExposeAddresses"),
		)

		opts := &types.GetDeploymentOption{
			Owner:        cctx.String("owner"),
			State:        []types.DeploymentState{types.DeploymentStateActive},
			DeploymentID: types.DeploymentID(cctx.String("id")),
		}

		if cctx.Bool("show-all") {
			opts.State = types.AllDeploymentStates
		}

		deployments, err := api.GetDeploymentList(ctx, opts)
		if err != nil {
			return err
		}

		for _, deployment := range deployments {
			for _, service := range deployment.Services {
				m := map[string]interface{}{
					"ID":              deployment.ID,
					"Name":            deployment.Name,
					"Image":           service.Image,
					"State":           types.DeploymentStateString(deployment.State),
					"CPU":             service.CPU,
					"Memory":          units.BytesSize(float64(service.Memory * units.MiB)),
					"Storage":         units.BytesSize(float64(service.Storage * units.MiB)),
					"CreatedTime":     deployment.CreatedAt.Format(defaultDateTimeLayout),
					"ExposeAddresses": fmt.Sprintf("%s:%d", deployment.ProviderExposeIP, service.Port),
				}
				tw.Write(m)
			}
		}

		tw.Flush(os.Stdout)
		return nil
	},
}

var DeleteDeployment = &cli.Command{
	Name:  "delete",
	Usage: "delete deployment",
	Flags: []cli.Flag{},
	Action: func(cctx *cli.Context) error {
		if cctx.NArg() != 1 {
			return IncorrectNumArgs(cctx)
		}

		api, closer, err := GetManagerAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()

		ctx := ReqContext(cctx)
		deploymentID := types.DeploymentID(cctx.Args().First())

		deployments, err := api.GetDeploymentList(ctx, &types.GetDeploymentOption{
			DeploymentID: deploymentID,
		})
		if err != nil {
			return err
		}

		if len(deployments) == 0 {
			return errors.New("deployment not found")
		}

		for _, deployment := range deployments {
			err = api.CloseDeployment(ctx, deployment)
			if err != nil {
				log.Errorf("delete deployment failed: %v", err)
			}
		}

		return nil
	},
}
