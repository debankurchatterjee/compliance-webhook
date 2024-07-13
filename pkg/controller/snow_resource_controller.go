package controller

import (
	"context"
	"github.com/compliance-webhook/internal/k8s"
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
	Get(ctx context.Context, name, namespace, operation string) (bool, error)
	Create(ctx context.Context, name, namespace, operation string) error
	Update(ctx context.Context, name, namespace, operation string) error
	Delete(ctx context.Context, name, namespace, operation string) error
}

func (s SnowResource) Get(ctx context.Context, name, namespace, operation string) (bool, error) {
	obj, err := s.DynamicKubernetesClient.Get(ctx, name, "", schema.GroupVersionResource{
		Group:    s.Group,
		Version:  s.Version,
		Resource: s.Resource,
	})
	if err != nil {
		return false, err
	}
	snowResource := obj.Object
	if status, ok := snowResource["status"]; ok {
		statusMap := status.(map[string]interface{})
		if val, ok1 := statusMap["OverallStatus"]; ok1 && val == "APPROVED" {
			return true, nil
		}
	}
	return false, err
}

func (s SnowResource) Create(ctx context.Context, name, namespace, operation string) error {
	//TODO implement me
	panic("implement me")
}

func (s SnowResource) Update(ctx context.Context, name, namespace, operation string) error {
	//TODO implement me
	panic("implement me")
}

func (s SnowResource) Delete(ctx context.Context, name, namespace, operation string) error {
	//TODO implement me
	panic("implement me")
}
