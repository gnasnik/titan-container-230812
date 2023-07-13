package provider

import (
	"fmt"
	"strings"

	"github.com/gnasnik/titan-container/api/types"
	"github.com/gnasnik/titan-container/node/impl/provider/kube/builder"
	"github.com/gnasnik/titan-container/node/impl/provider/kube/manifest"
)

const (
	podReplicas = 1
)

func ClusterDeploymentFromDeployment(deployment *types.Deployment) (builder.IClusterDeployment, error) {
	if len(deployment.ID) == 0 {
		return nil, fmt.Errorf("deployment ID can not empty")
	}

	deploymentID := manifest.DeploymentID{ID: string(deployment.ID), Owner: deployment.Owner}
	group, err := deploymentToManifestGroup(deployment)
	if err != nil {
		return nil, err
	}

	settings := builder.ClusterSettings{
		SchedulerParams: make([]*builder.SchedulerParams, len(group.Services)),
	}

	return &builder.ClusterDeployment{
		Did:     deploymentID,
		Group:   group,
		Sparams: settings,
	}, nil
}

func deploymentToManifestGroup(deployment *types.Deployment) (*manifest.Group, error) {
	if len(deployment.Services) == 0 {
		return nil, fmt.Errorf("deployment service can not empty")
	}

	services := make([]manifest.Service, 0, len(deployment.Services))
	for _, service := range deployment.Services {
		s, err := serviceToManifestService(service)
		if err != nil {
			return nil, err
		}
		services = append(services, s)
	}

	return &manifest.Group{Services: services}, nil
}

func serviceToManifestService(service *types.Service) (manifest.Service, error) {
	if len(service.Image) == 0 {
		return manifest.Service{}, fmt.Errorf("service image can not empty")
	}
	name := imageToServiceName(service.Image)
	resource := resourceToManifestResource(&service.ComputeResources)
	expose := exposeFromPort(service.Port)
	s := manifest.Service{
		Name:      name,
		Image:     service.Image,
		Resources: &resource,
		Expose:    make([]*manifest.ServiceExpose, 0),
		Count:     podReplicas,
	}

	if expose != nil {
		s.Expose = append(s.Expose, expose)
	}

	return s, nil
}

func imageToServiceName(image string) string {
	names := strings.Split(image, ":")
	if len(names) > 0 {
		return names[0]
	}
	return image
}

func resourceToManifestResource(resource *types.ComputeResources) manifest.ResourceUnits {
	return *manifest.NewResourceUnits(uint64(resource.CPU*1000), uint64(resource.Memory*1000000), uint64(resource.Storage*1000000))
}

func exposeFromPort(port int) *manifest.ServiceExpose {
	if port == 0 {
		return nil
	}
	return &manifest.ServiceExpose{Port: uint32(port), ExternalPort: uint32(port), Proto: manifest.TCP, Global: true}
}
