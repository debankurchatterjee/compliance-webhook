package k8s

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

type KubernetesCustomResourceClient struct {
	DynamicClient dynamic.DynamicClient
}

const snowLabel = "snow.controller/changeID"

func NewKubernetesCustomResourceClient(inClusterConfig bool) (*KubernetesCustomResourceClient, error) {
	var dynamicClient *dynamic.DynamicClient
	if inClusterConfig {
		config, err := rest.InClusterConfig()
		dynamicClient, err = dynamic.NewForConfig(config)
		if err != nil {
			return nil, err
		}
	} else {
		var kubeconfig string
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
		dynamicClient, err = dynamic.NewForConfig(config)
		if err != nil {
			return nil, err
		}
	}
	return &KubernetesCustomResourceClient{
		DynamicClient: *dynamicClient,
	}, nil
}

type KubernetesClient interface {
	Get(ctx context.Context, name string, namespace string,
		gvr schema.GroupVersionResource) (*unstructured.Unstructured, error)
	Create(ctx context.Context, payload *unstructured.Unstructured,
		gvr schema.GroupVersionResource) (*unstructured.Unstructured, error)
}

func (c *KubernetesCustomResourceClient) Get(ctx context.Context, name string, namespace string,
	gvr schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	labelSelector := fmt.Sprintf("%s=%s", snowLabel, name)
	list, err := c.DynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}
	if len(list.Items) == 0 {
		return nil, fmt.Errorf("resource not found with label %s=%s", snowLabel, name)
	}
	return &list.Items[0], nil
}

func (c *KubernetesCustomResourceClient) Create(ctx context.Context,
	payload *unstructured.Unstructured,
	gvr schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	return c.DynamicClient.Resource(gvr).Create(ctx, payload, metav1.CreateOptions{})
}
