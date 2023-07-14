package kube

import (
	"context"
	"fmt"
	"os"

	"github.com/gnasnik/titan-container/node/impl/provider/kube/builder"
	logging "github.com/ipfs/go-log/v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/flowcontrol"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
)

type Client interface {
	Deploy(ctx context.Context, deployment builder.IClusterDeployment) error
	DeleteNS(ctx context.Context, ns string) error
	FetchNodeResources(ctx context.Context) (map[string]*nodeResource, error)
	ListServices(ctx context.Context, ns string) (*corev1.ServiceList, error)
	ListDeployments(ctx context.Context, ns string) (*appsv1.DeploymentList, error)
	GetNS(ctx context.Context, ns string) (*v1.Namespace, error)
}

type client struct {
	kc   kubernetes.Interface
	metc metricsclient.Interface
	log  *logging.ZapEventLogger
}

func openKubeConfig(cfgPath string) (*rest.Config, error) {
	// Always bypass the default rate limiting
	rateLimiter := flowcontrol.NewFakeAlwaysRateLimiter()

	if cfgPath != "" {
		cfgPath = os.ExpandEnv(cfgPath)

		if _, err := os.Stat(cfgPath); err == nil {
			cfg, err := clientcmd.BuildConfigFromFlags("", cfgPath)
			if err != nil {
				return cfg, fmt.Errorf("%w: error building kubernetes config", err)
			}
			cfg.RateLimiter = rateLimiter
			return cfg, err
		}
	}

	cfg, err := rest.InClusterConfig()
	if err != nil {
		return cfg, fmt.Errorf("%w: error building kubernetes config", err)
	}
	cfg.RateLimiter = rateLimiter

	return cfg, err
}

func NewClient(configPath string) (Client, error) {
	config, err := openKubeConfig(configPath)
	if err != nil {
		return nil, err
	}

	// create the clientSet
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	metc, err := metricsclient.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	var log = logging.Logger("client")

	return &client{kc: clientSet, metc: metc, log: log}, nil
}

func (c *client) Deploy(ctx context.Context, deployment builder.IClusterDeployment) error {
	// lid := cdeployment.LeaseID()
	group := deployment.ManifestGroup()

	settingsI := ctx.Value(builder.SettingsKey)
	if nil == settingsI {
		return fmt.Errorf("kube client: not configured with settings in the context passed to function")
	}
	settings := settingsI.(builder.Settings)
	if err := builder.ValidateSettings(settings); err != nil {
		return err
	}

	ns := builder.BuildNS(settings, deployment)
	if err := applyNS(ctx, c.kc, builder.BuildNS(settings, deployment)); err != nil {
		c.log.Errorf("applying namespace %s err %s", ns.Name(), err.Error())
		return err
	}

	if err := applyNetPolicies(ctx, c.kc, builder.BuildNetPol(settings, deployment)); err != nil { //
		c.log.Errorf("applying namespace %s network policies err %s", ns.Name(), err)
		return err
	}

	// cmanifest := builder.BuildManifest(c.log, settings, c.ns, cdeployment)
	// if err := applyManifest(ctx, c.ac, cmanifest); err != nil {
	// 	c.log.Error("applying manifest", "err", err, "lease", lid)
	// 	return err
	// }

	// if err := cleanupStaleResources(ctx, c.kc, lid, group); err != nil {
	// 	c.log.Error("cleaning stale resources", "err", err, "lease", lid)
	// 	return err
	// }

	for svcIdx := range group.Services {
		workload := builder.NewWorkload(settings, deployment, svcIdx)

		service := &group.Services[svcIdx]

		persistent := false
		for i := range service.Resources.Storage {
			attrVal := service.Resources.Storage[i].Attributes.Find(builder.StorageClassDefault)
			if persistent, _ = attrVal.AsBool(); persistent {
				break
			}
		}

		if persistent {
			if err := applyStatefulSet(ctx, c.kc, builder.BuildStatefulSet(workload)); err != nil {
				c.log.Errorf("applying statefulSet err %s, ns %s, service %s", err.Error(), ns.Name(), service.Name)
				return err
			}
		} else {
			if err := applyDeployment(ctx, c.kc, builder.NewDeployment(workload)); err != nil {
				c.log.Errorf("applying deployment err %s, ns %s, service %s", err.Error(), ns.Name(), service.Name)
				return err
			}
		}

		if len(service.Expose) == 0 {
			c.log.Debug("no services", "ns", ns.Name(), "service", service.Name)
			continue
		}

		serviceBuilderLocal := builder.BuildService(workload, false)
		if serviceBuilderLocal.Any() {
			if err := applyService(ctx, c.kc, serviceBuilderLocal); err != nil {
				c.log.Error("applying local service err %s, ns %s, service %s", err.Error(), ns.Name(), service.Name)
				return err
			}
		}

		serviceBuilderGlobal := builder.BuildService(workload, true)
		if serviceBuilderGlobal.Any() {
			if err := applyService(ctx, c.kc, serviceBuilderGlobal); err != nil {
				c.log.Error("applying global service err %s, ns %s, service %s", err.Error(), ns.Name(), service.Name)
				return err
			}
		}
	}

	return nil
}

func (c *client) DeleteNS(ctx context.Context, ns string) error {
	return c.kc.CoreV1().Namespaces().Delete(ctx, ns, metav1.DeleteOptions{})
}

func (c *client) GetNS(ctx context.Context, ns string) (*v1.Namespace, error) {
	return c.kc.CoreV1().Namespaces().Get(ctx, ns, metav1.GetOptions{})
}

// func (c *client) ListNodes(ctx context.Context) (*corev1.NodeList, error) {
// 	return c.kc.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
// }

func (c *client) ListServices(ctx context.Context, ns string) (*corev1.ServiceList, error) {
	return c.kc.CoreV1().Services(ns).List(ctx, metav1.ListOptions{})
}

func (c *client) ListDeployments(ctx context.Context, ns string) (*appsv1.DeploymentList, error) {
	return c.kc.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{})
}

func (c *client) ListPods(ctx context.Context, ns string) (*corev1.PodList, error) {
	return c.kc.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
}
