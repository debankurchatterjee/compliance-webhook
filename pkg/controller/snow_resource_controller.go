package controller

import (
	"context"
	"fmt"
	"github.com/compliance-webhook/internal/k8s"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// SnowResource struct which wraps around k8s dynamic client to create snow resource.
type SnowResource struct {
	DynamicKubernetesClient k8s.KubernetesClient
	Group                   string
	Version                 string
	Resource                string
}

// NewSnowResource is a constructor for SnowResource
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

// SnowResourceController is the interface to handle GET and CREATE operation on snow resource
type SnowResourceController interface {
	Get(ctx context.Context, label, namespace, operation string, bypassStatusCheck bool) (string, bool, error)
	Create(ctx context.Context, name, namespace, operation, kind, payload string, labels map[string]string, generateName bool) (string, error)
}

// Get will get snow resource based on labels
func (s SnowResource) Get(ctx context.Context, label, namespace, operation string, bypassStatusCheck bool) (name string, isAvailable bool, err1 error) {
	var obj *unstructured.Unstructured
	var err error
	if operation == "update" {
		obj, err = s.DynamicKubernetesClient.GetLatest(ctx, label, "", schema.GroupVersionResource{
			Group:    s.Group,
			Version:  s.Version,
			Resource: s.Resource,
		})
	} else {
		obj, err = s.DynamicKubernetesClient.Get(ctx, label, "", schema.GroupVersionResource{
			Group:    s.Group,
			Version:  s.Version,
			Resource: s.Resource,
		})
	}
	if err != nil {
		return "", false, err
	}
	if bypassStatusCheck && obj != nil {
		return obj.GetName(), true, nil
	}
	snowResource := obj.Object
	if status, ok := snowResource["status"]; ok {
		statusMap, ok := status.(map[string]interface{})
		if !ok {
			return "", false, fmt.Errorf("unable to parse status subresource from cr")
		}
		if val, ok1 := statusMap["OverallStatus"]; ok1 && val == "APPROVED" {
			return obj.GetName(), true, nil
		}
	}
	return "", false, err
}

// Create will create snow resource based on below arguments
func (s SnowResource) Create(ctx context.Context, name, namespace, operation, kind, payload string, labels map[string]string, generateName bool) (string, error) {
	createName := fmt.Sprintf("%s-%s-%s", name, namespace, operation)
	if operation == "update" {
		createName = fmt.Sprintf("%s-1", createName)
	}
	if !generateName {
		createName = name
	}
	obj := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": fmt.Sprintf("%s/%s", s.Group, s.Version),
			"kind":       "Snow",
			"metadata": map[string]interface{}{
				"name": createName,
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
	return createName, err
}
