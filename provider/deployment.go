package provider

import (
	"strings"

	"github.com/gnasnik/titan-container/api/types"
	"github.com/gnasnik/titan-container/provider/kube/builder"
	"github.com/gnasnik/titan-container/provider/kube/manifest"
)

const (
	podReplicas = 1
)

func ClusterDeploymentFromDeployment(deployment *types.Deployment) builder.IClusterDeployment {
	deploymentID := manifest.DeploymentID{ID: string(deployment.ID), Owner: deployment.Owner}
	group := deploymentToManifestGroup(deployment)

	settings := builder.ClusterSettings{
		SchedulerParams: make([]*builder.SchedulerParams, len(group.Services)),
	}

	return &builder.ClusterDeployment{
		Did:     deploymentID,
		Group:   group,
		Sparams: settings,
	}
}

func deploymentToManifestGroup(deployment *types.Deployment) *manifest.Group {
	services := make([]manifest.Service, 0, len(deployment.Services))
	for _, service := range deployment.Services {
		s := serviceToManifestService(service)
		services = append(services, s)
	}

	return &manifest.Group{Services: services}
}

func serviceToManifestService(service *types.Service) manifest.Service {
	name := imageToServiceName(service.Image)
	resource := resourceToManifestResource(&service.ComputeResources)
	expose := exposeFromPort(service.Port)
	return manifest.Service{
		Name:      name,
		Image:     service.Image,
		Resources: &resource,
		Expose:    []*manifest.ServiceExpose{&expose},
		Count:     podReplicas,
	}
}

func imageToServiceName(image string) string {
	names := strings.Split(image, ":")
	if len(names) > 0 {
		return names[0]
	}
	return image
}

func resourceToManifestResource(resource *types.ComputeResources) manifest.ResourceUnits {
	return *manifest.NewResourceUnits(uint64(resource.CPU*1000), uint64(resource.Memory*1000000))
}

func exposeFromPort(port int) manifest.ServiceExpose {
	return manifest.ServiceExpose{Port: uint32(port), ExternalPort: uint32(port), Proto: manifest.TCP, Global: true}
}
