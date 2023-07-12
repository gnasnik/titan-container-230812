package kube

// nolint:deadcode,golint

import (
	"context"

	"github.com/gnasnik/titan-container/provider/kube/builder"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func applyNS(ctx context.Context, kc kubernetes.Interface, b builder.NS) error {
	obj, err := kc.CoreV1().Namespaces().Get(ctx, b.Name(), metav1.GetOptions{})

	switch {
	case err == nil:
		obj, err = b.Update(obj)
		if err == nil {
			_, err = kc.CoreV1().Namespaces().Update(ctx, obj, metav1.UpdateOptions{})
		}
	case errors.IsNotFound(err):
		obj, err = b.Create()
		if err == nil {
			_, err = kc.CoreV1().Namespaces().Create(ctx, obj, metav1.CreateOptions{})
		}
	}
	return err
}

// Apply list of Network Policies
func applyNetPolicies(ctx context.Context, kc kubernetes.Interface, b builder.NetPol) error {
	var err error

	policies, err := b.Create()
	if err != nil {
		return err
	}

	for _, pol := range policies {
		obj, err := kc.NetworkingV1().NetworkPolicies(b.NS()).Get(ctx, pol.Name, metav1.GetOptions{})

		switch {
		case err == nil:
			_, err = b.Update(obj)
			if err == nil {
				_, err = kc.NetworkingV1().NetworkPolicies(b.NS()).Update(ctx, pol, metav1.UpdateOptions{})
			}
		case errors.IsNotFound(err):
			_, err = kc.NetworkingV1().NetworkPolicies(b.NS()).Create(ctx, pol, metav1.CreateOptions{})
		}
		if err != nil {
			break
		}
	}

	return err
}

func applyDeployment(ctx context.Context, kc kubernetes.Interface, b builder.Deployment) error {
	obj, err := kc.AppsV1().Deployments(b.NS()).Get(ctx, b.Name(), metav1.GetOptions{})

	switch {
	case err == nil:
		obj, err = b.Update(obj)

		if err == nil {
			_, err = kc.AppsV1().Deployments(b.NS()).Update(ctx, obj, metav1.UpdateOptions{})

		}
	case errors.IsNotFound(err):
		obj, err = b.Create()
		if err == nil {
			_, err = kc.AppsV1().Deployments(b.NS()).Create(ctx, obj, metav1.CreateOptions{})
		}
	}
	return err
}

func applyStatefulSet(ctx context.Context, kc kubernetes.Interface, b builder.StatefulSet) error {
	obj, err := kc.AppsV1().StatefulSets(b.NS()).Get(ctx, b.Name(), metav1.GetOptions{})

	switch {
	case err == nil:
		obj, err = b.Update(obj)

		if err == nil {
			_, err = kc.AppsV1().StatefulSets(b.NS()).Update(ctx, obj, metav1.UpdateOptions{})

		}
	case errors.IsNotFound(err):
		obj, err = b.Create()
		if err == nil {
			_, err = kc.AppsV1().StatefulSets(b.NS()).Create(ctx, obj, metav1.CreateOptions{})
		}
	}
	return err
}

func applyService(ctx context.Context, kc kubernetes.Interface, b builder.Service) error {
	obj, err := kc.CoreV1().Services(b.NS()).Get(ctx, b.Name(), metav1.GetOptions{})

	switch {
	case err == nil:
		obj, err = b.Update(obj)
		if err == nil {
			_, err = kc.CoreV1().Services(b.NS()).Update(ctx, obj, metav1.UpdateOptions{})
		}
	case errors.IsNotFound(err):
		obj, err = b.Create()
		if err == nil {
			_, err = kc.CoreV1().Services(b.NS()).Create(ctx, obj, metav1.CreateOptions{})
		}
	}
	return err
}

// func applyManifest(ctx context.Context, kc crdapi.Interface, b builder.Manifest) error {
// 	obj, err := kc.AkashV2beta2().Manifests(b.NS()).Get(ctx, b.Name(), metav1.GetOptions{})

// 	switch {
// 	case err == nil:
// 		// TODO - only run this update if it would change something
// 		obj, err = b.Update(obj)
// 		if err == nil {
// 			_, err = kc.AkashV2beta2().Manifests(b.NS()).Update(ctx, obj, metav1.UpdateOptions{})
// 		}
// 	case errors.IsNotFound(err):
// 		obj, err = b.Create()
// 		if err == nil {
// 			_, err = kc.AkashV2beta2().Manifests(b.NS()).Create(ctx, obj, metav1.CreateOptions{})
// 		}
// 	}
// 	return err
// }
