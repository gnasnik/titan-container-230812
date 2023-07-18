package cli

import (
	"encoding/json"
	"fmt"
	"github.com/docker/go-units"
	"github.com/gnasnik/titan-container/api/types"
	"github.com/gnasnik/titan-container/lib/tablewriter"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"os"
)

var defaultDateTimeLayout = "2006-01-02 15:04:05"

var deploymentCmds = &cli.Command{
	Name:  "deployment",
	Usage: "Manager deployment",
	Subcommands: []*cli.Command{
		CreateDeployment,
		DeploymentList,
		DeleteDeployment,
		StatusDeployment,
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
		&cli.IntFlag{
			Name:  "port",
			Usage: "deployment internal server port",
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
		&cli.StringFlag{
			Name:  "env",
			Usage: "set the deployment running environment",
		},
		&cli.StringSliceFlag{
			Name:  "args",
			Usage: "set the deployment running arguments",
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

		var env types.Env
		if cctx.String("env") != "" {
			err := json.Unmarshal([]byte(cctx.String("env")), &env)
			if err != nil {
				return err
			}
		}

		deployment := &types.Deployment{
			ProviderID: providerID,
			Name:       cctx.String("name"),
			Services: []*types.Service{
				{
					Image: cctx.String("image"),
					Port:  cctx.Int("port"),
					ComputeResources: types.ComputeResources{
						CPU:     cctx.Float64("cpu"),
						Memory:  cctx.Int64("mem"),
						Storage: cctx.Int64("storage"),
					},
					Env:       env,
					Arguments: cctx.StringSlice("args"),
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
			tablewriter.Col("Image"),
			tablewriter.Col("State"),
			tablewriter.Col("Total"),
			tablewriter.Col("Ready"),
			tablewriter.Col("Available"),
			tablewriter.Col("CPU"),
			tablewriter.Col("Memory"),
			tablewriter.Col("Storage"),
			tablewriter.Col("Provider"),
			tablewriter.Col("Port"),
			tablewriter.Col("CreatedTime"),
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
				state := types.DeploymentStateInActive
				if service.Status.TotalReplicas == service.Status.ReadyReplicas {
					state = types.DeploymentStateActive
				}

				m := map[string]interface{}{
					"ID":          deployment.ID,
					"Image":       service.Image,
					"State":       types.DeploymentStateString(state),
					"Total":       service.Status.TotalReplicas,
					"Ready":       service.Status.ReadyReplicas,
					"Available":   service.Status.AvailableReplicas,
					"CPU":         service.CPU,
					"Memory":      units.BytesSize(float64(service.Memory * units.MiB)),
					"Storage":     units.BytesSize(float64(service.Storage * units.MiB)),
					"Provider":    deployment.ProviderExposeIP,
					"Port":        fmt.Sprintf("%d->%d", service.Port, service.ExposePort),
					"CreatedTime": deployment.CreatedAt.Format(defaultDateTimeLayout),
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

var StatusDeployment = &cli.Command{
	Name:  "status",
	Usage: "show deployment status",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "log",
			Usage: "show deployment log",
		},
	},
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

		var deployment *types.Deployment
		for _, d := range deployments {
			if d.ID == deploymentID {
				deployment = d
				continue
			}
		}

		if deployment == nil {
			return errors.New("deployment not found")
		}

		fmt.Printf("DeploymentID:\t%s\n", deployment.ID)
		fmt.Printf("State:\t\t%s\n", types.DeploymentStateString(deployment.State))
		fmt.Printf("CreadTime:\t%v\n", deployment.CreatedAt)
		fmt.Printf("--------\nEvents:\n")

		serviceEvents, err := api.GetEvents(ctx, deployment)
		if err != nil {
			return err
		}

		for _, sv := range serviceEvents {
			for i, event := range sv.Events {
				fmt.Printf("%d.\t[%s]\t%s\n", i, sv.ServiceName, event)
			}
		}

		if cctx.Bool("log") {
			fmt.Printf("--------\nLogs:\n")

			serviceLogs, err := api.GetLogs(ctx, deployment)
			if err != nil {
				return err
			}

			for _, sl := range serviceLogs {
				for _, l := range sl.Logs {
					fmt.Printf("%s\n", l)
				}
			}
		}

		return nil
	},
}
