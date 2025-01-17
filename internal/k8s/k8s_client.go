package k8s

import (
	"context"
	"fmt"
	"github.com/compliance-webhook/internal/logutil/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

// KubernetesCustomResourceClient which defines the dynamic client to do custom resource LCM
type KubernetesCustomResourceClient struct {
	DynamicClient dynamic.DynamicClient
}

const snowLabel = "snow.controller/changeID"
const parentLabel = "snow.controller/parentChangeID"
const snowNamespace = "snow-compliance"

// NewKubernetesCustomResourceClient is the constructor for KubernetesCustomResourceClient
func NewKubernetesCustomResourceClient(inClusterConfig bool) (*KubernetesCustomResourceClient, error) {
	var dynamicClient *dynamic.DynamicClient
	if inClusterConfig {
		config, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
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

// KubernetesClient is the interface to implement KubernetesCustomResourceClient for GET and CREATE Methods
type KubernetesClient interface {
	Get(ctx context.Context, label string, namespace string,
		gvr schema.GroupVersionResource) (*unstructured.Unstructured, error)
	Create(ctx context.Context, payload *unstructured.Unstructured,
		gvr schema.GroupVersionResource) (*unstructured.Unstructured, error)
	GetLatest(ctx context.Context, label string, namespace string,
		gvr schema.GroupVersionResource) (*unstructured.Unstructured, error)
}

// Get method will fetch the initial result based on the label selector
// this is mainly used for owner reference if one of the owner is approved then request will be auto approved
func (c *KubernetesCustomResourceClient) Get(ctx context.Context, label string, namespace string,
	gvr schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	logger := log.From(ctx)
	logger.Info("context labels", "Label", label)
	labelSelector := fmt.Sprintf("%s=%s", snowLabel, label)
	list, err := c.DynamicClient.Resource(gvr).Namespace(snowNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}
	if len(list.Items) == 0 {
		return nil, fmt.Errorf("resource not found with label %s=%s", parentLabel, label)
	}
	return &list.Items[0], nil
}

// GetLatest method will fetch the latest result based on label selector this method will help to determine last updated resource
func (c *KubernetesCustomResourceClient) GetLatest(ctx context.Context, label string, namespace string,
	gvr schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	logger := log.From(ctx)
	logger.Info("context labels", "Label", label)
	labelSelector := fmt.Sprintf("%s=%s", parentLabel, label)
	list, err := c.DynamicClient.Resource(gvr).Namespace(snowNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}
	if len(list.Items) == 0 {
		return nil, fmt.Errorf("resource not found with label %s=%s", snowLabel, label)
	}
	return &list.Items[len(list.Items)-1], nil
}

// Create will create the CR
func (c *KubernetesCustomResourceClient) Create(ctx context.Context,
	payload *unstructured.Unstructured,
	gvr schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	client := c.DynamicClient.Resource(gvr).Namespace(snowNamespace)
	create, err := client.Create(ctx, payload, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return create, nil
}
