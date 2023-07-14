package provider

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gnasnik/titan-container/api/types"
	"github.com/gnasnik/titan-container/node/impl/provider/kube/builder"
	"github.com/gnasnik/titan-container/node/impl/provider/kube/manifest"
	"github.com/google/uuid"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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
		Args:      service.Arguments,
		Env:       envToManifestEnv(service.Env),
		Resources: &resource,
		Expose:    make([]*manifest.ServiceExpose, 0),
		Count:     podReplicas,
	}

	if expose != nil {
		s.Expose = append(s.Expose, expose)
	}

	return s, nil
}

func envToManifestEnv(serviceEnv types.Env) []string {
	envs := make([]string, 0, len(serviceEnv))
	for k, v := range serviceEnv {
		env := fmt.Sprintf("%s=%s", k, v)
		envs = append(envs, env)
	}
	return envs
}

func imageToServiceName(image string) string {
	serviceName := image
	names := strings.Split(image, ":")
	if len(names) > 0 {
		serviceName = names[0]
	}

	uuidString := uuid.NewString()
	uuidString = strings.Replace(uuidString, "-", "", -1)

	return fmt.Sprintf("%s-%s", serviceName, uuidString)
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

func k8sDeploymentsToServices(deploymentList *appsv1.DeploymentList) ([]*types.Service, error) {
	services := make([]*types.Service, 0, len(deploymentList.Items))

	for _, deployment := range deploymentList.Items {
		s, err := k8sDeploymentToService(&deployment)
		if err != nil {
			return nil, err
		}
		services = append(services, s)
	}

	return services, nil
}

func k8sDeploymentToService(deployment *appsv1.Deployment) (*types.Service, error) {
	if len(deployment.Spec.Template.Spec.Containers) == 0 {
		return nil, fmt.Errorf("deployment container can not empty")
	}

	container := deployment.Spec.Template.Spec.Containers[0]
	service := &types.Service{Image: container.Image, Name: container.Name}
	service.CPU = container.Resources.Limits.Cpu().AsApproximateFloat64()
	service.Memory = container.Resources.Limits.Memory().Value() / 1000000
	service.Storage = int64(container.Resources.Limits.StorageEphemeral().AsApproximateFloat64()) / 1000000
	if len(container.Ports) > 0 {
		service.Port = int(container.Ports[0].ContainerPort)
	}

	if len(deployment.Status.Conditions) == 0 {
		return nil, fmt.Errorf("deployment conditions can not empty")
	}

	conditions := deployment.Status.Conditions
	sort.Slice(conditions, func(i, j int) bool {
		return conditions[i].LastUpdateTime.Before(&conditions[j].LastUpdateTime)
	})

	lastCondition := conditions[len(conditions)-1]
	service.State = getConditionStatus(lastCondition)
	service.ErrorMessage = lastCondition.Message

	return service, nil
}

func k8sServiceToPortMap(serviceList *corev1.ServiceList) (map[string]int, error) {
	portMap := make(map[string]int)
	for _, service := range serviceList.Items {
		if len(service.Spec.Ports) > 0 {
			servicePort := service.Spec.Ports[0]
			if servicePort.NodePort != 0 {
				serviceName := strings.TrimSuffix(service.Name, builder.SuffixForNodePortServiceName)
				portMap[serviceName] = int(servicePort.NodePort)
			} else {
				portMap[service.Name] = int(servicePort.Port)
			}
		}

	}
	return portMap, nil
}

func getConditionStatus(condition appsv1.DeploymentCondition) types.ServiceState {
	switch condition.Status {
	case corev1.ConditionTrue:
		return types.ServiceStateNormal
	case corev1.ConditionFalse:
		return types.ServiceStateError
	}
	return types.ServiceStateUnknown
}
