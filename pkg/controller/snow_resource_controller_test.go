package controller

import (
	"context"
	"github.com/compliance-webhook/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func getSnowResourceControllerMock(t *testing.T) *mock.MockSnowResourceController {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	return mock.NewMockSnowResourceController(ctrl)
}

func TestSnowResource_Create(t *testing.T) {
	mockClient := getSnowResourceControllerMock(t)
	labels := make(map[string]string)
	labels["snow.controller/changeID"] = "eq2u23i234f223223fj2kfjervervre"
	mockClient.EXPECT().Create(context.Background(), "test", "snow-compliance", "create", "deployment", "", labels, true).Return("", nil)
	_, err := mockClient.Create(context.Background(), "test", "snow-compliance", "create", "deployment", "", labels, true)
	if err != nil {
		return
	}
	assert.NoError(t, err)
}

func TestSnowResource_Get(t *testing.T) {
	mockClient := getSnowResourceControllerMock(t)
	mockClient.EXPECT().Get(context.Background(), "eq2u23i234f223223fj2kfjervervre", "test", "create", true).Return("", false, nil)
	_, _, err := mockClient.Get(context.Background(), "eq2u23i234f223223fj2kfjervervre", "test", "create", true)
	if err != nil {
		return
	}
	assert.NoError(t, err)
}
