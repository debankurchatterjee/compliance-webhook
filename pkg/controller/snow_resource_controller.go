package controller

import (
	"context"
	"fmt"
	"github.com/compliance-webhook/internal/k8s"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type SnowResource struct {
	DynamicKubernetesClient k8s.KubernetesClient
	Group                   string
	Version                 string
	Resource                string
}

func NewSnowResource(group, version, resource string, isInClusterConfig bool) (SnowResource, error) {
	k8sDynamicClient, err := k8s.NewKubernetesCustomResourceClient(isInClusterConfig)
	if err != nil {
		return SnowResource{}, nil
	}
	return SnowResource{
		DynamicKubernetesClient: k8sDynamicClient,
		Group:                   group,
		Version:                 version,
		Resource:                resource,
	}, nil
}

type SnowResourceController interface {
	Get(ctx context.Context, label, namespace, operation string, bypassStatusCheck bool) (bool, error)
	Create(ctx context.Context, name, namespace, operation, kind, payload string, labels map[string]string) error
	Update(ctx context.Context, name, namespace, operation string) error
	Delete(ctx context.Context, name, namespace, operation string) error
}

func (s SnowResource) Get(ctx context.Context, label, namespace, operation string, bypassStatusCheck bool) (bool, error) {
	obj, err := s.DynamicKubernetesClient.Get(ctx, label, "", schema.GroupVersionResource{
		Group:    s.Group,
		Version:  s.Version,
		Resource: s.Resource,
	})
	if err != nil {
		return false, err
	}
	if bypassStatusCheck && obj != nil {
		return true, nil
	}
	snowResource := obj.Object
	if status, ok := snowResource["status"]; ok {
		statusMap, ok := status.(map[string]interface{})
		if !ok {
			return false, fmt.Errorf("unable to parse status subresource from cr")
		}
		if val, ok1 := statusMap["OverallStatus"]; ok1 && val == "APPROVED" {
			return true, nil
		}
	}
	return false, err
}

func (s SnowResource) Create(ctx context.Context, name, namespace, operation, kind, payload string, labels map[string]string) error {
	obj := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": fmt.Sprintf("%s/%s", s.Group, s.Version),
			"kind":       "Snow",
			"metadata": map[string]interface{}{
				"name": fmt.Sprintf("%s-%s", name, namespace),
			},
			"spec": map[string]interface{}{
				"operation":  operation,
				"changeName": name,
				"namespace":  namespace,
				"kind":       kind,
				"payload":    payload,
			},
		},
	}
	obj.SetLabels(labels)
	_, err := s.DynamicKubernetesClient.Create(ctx, &obj, schema.GroupVersionResource{
		Group:    s.Group,
		Version:  s.Version,
		Resource: s.Resource,
	})
	return err
}

func (s SnowResource) Update(ctx context.Context, name, namespace, operation string) error {
	return nil
}

func (s SnowResource) Delete(ctx context.Context, name, namespace, operation string) error {
	return nil
}
