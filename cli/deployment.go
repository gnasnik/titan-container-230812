package cli

import (
	"fmt"
	"github.com/gnasnik/titan-container/api/types"
	"github.com/urfave/cli/v2"
)

var deploymentCmds = &cli.Command{
	Name:  "deployment",
	Usage: "Manager deployment",
	Subcommands: []*cli.Command{
		CreateDeployment,
		DeploymentList,
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
			ID:         types.DeploymentID(cctx.String("deploy-id")),
			ProviderID: providerID,
			Name:       cctx.String("name"),
			Image:      cctx.String("image"),
			Env: map[string]string{
				"hello": "test",
			},
			Services: []*types.Service{
				{
					Image: cctx.String("image"),
					ComputeResources: types.ComputeResources{
						CPU:     cctx.Float64("cpu"),
						Memory:  cctx.Int64("mem"),
						Storage: cctx.Int64("storage"),
					},
				},
			},
		}

		return api.CreateDeployment(ctx, providerID, deployment)
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
	},
	Action: func(cctx *cli.Context) error {
		api, closer, err := GetManagerAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()

		ctx := ReqContext(cctx)

		opts := &types.GetDeploymentOption{
			Owner:        cctx.String("owner"),
			DeploymentID: types.DeploymentID(cctx.String("id")),
		}

		deployments, err := api.GetDeploymentList(ctx, opts)
		if err != nil {
			return err
		}

		for _, deployment := range deployments {
			fmt.Println(deployment)
		}

		return nil
	},
}
