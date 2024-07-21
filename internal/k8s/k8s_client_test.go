package k8s

import (
	"context"
	"fmt"
	"github.com/compliance-webhook/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"
)

func TestKubernetesCustomResourceClient_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	// Create the mock client
	mockClient := mock.NewMockKubernetesClient(ctrl)
	obj := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": fmt.Sprintf("%s/%s", "", ""),
			"kind":       "Snow",
			"metadata": map[string]interface{}{
				"name": "test",
			},
			"spec": map[string]interface{}{
				"operation":  "create",
				"changeName": "test-1",
				"namespace":  "default",
				"kind":       "deployment",
				"payload":    "",
			},
		},
	}

	mockClient.EXPECT().Create(context.Background(), &obj, schema.GroupVersionResource{
		Group:    "compliance.complaince.org",
		Version:  "v1",
		Resource: "snows",
	}).Return(nil, nil)
	_, err := mockClient.Create(context.Background(), &obj, schema.GroupVersionResource{
		Group:    "compliance.complaince.org",
		Version:  "v1",
		Resource: "snows",
	})
	assert.NoError(t, err)
}

func TestKubernetesCustomResourceClient_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	// Create the mock client
	mockClient := mock.NewMockKubernetesClient(ctrl)
	mockClient.EXPECT().Get(context.Background(), "snow.controller/changeID=dqyru1i2rufqfqhifqw", "test", schema.GroupVersionResource{
		Group:    "compliance.complaince.org",
		Version:  "v1",
		Resource: "snows",
	}).Return(nil, nil)
	_, err := mockClient.Get(context.Background(), "snow.controller/changeID=dqyru1i2rufqfqhifqw", "test", schema.GroupVersionResource{
		Group:    "compliance.complaince.org",
		Version:  "v1",
		Resource: "snows"})
	assert.NoError(t, err)
}

func TestKubernetesCustomResourceClient_GetLatest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	// Create the mock client
	mockClient := mock.NewMockKubernetesClient(ctrl)
	mockClient.EXPECT().GetLatest(context.Background(), "snow.controller/parentID=dqyru1i2rufqfqhifqw", "test", schema.GroupVersionResource{
		Group:    "compliance.complaince.org",
		Version:  "v1",
		Resource: "snows",
	}).Return(nil, nil)
	_, err := mockClient.GetLatest(context.Background(), "snow.controller/parentID=dqyru1i2rufqfqhifqw", "test", schema.GroupVersionResource{
		Group:    "compliance.complaince.org",
		Version:  "v1",
		Resource: "snows"})
	assert.NoError(t, err)
}
