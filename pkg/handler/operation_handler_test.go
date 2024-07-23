package handler

import (
	"context"
	"github.com/compliance-webhook/internal/logutil/log"
	"github.com/compliance-webhook/mock"
	"github.com/compliance-webhook/pkg/controller"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	admissionv1 "k8s.io/api/admission/v1"
	"testing"
)

func Test_admissionOperationHandler_handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	admissionOperationHandler := mock.NewMockoperationHandlerFactory(ctrl)
	ctx := context.Background()
	logger := log.From(ctx)
	ops := admissionv1.Update
	snowController, err := controller.NewSnowResource(group, version, resource, false)
	if err != nil {
		return
	}
	admissionOperationHandler.EXPECT().Handle(ctx, &admissionv1.AdmissionRequest{}, &ops, snowController, "nginx-app", "", "deployment", []interface{}{}, logger).Return(nil, nil)
	_, err = admissionOperationHandler.Handle(ctx, &admissionv1.AdmissionRequest{}, &ops, snowController, "nginx-app", "", "deployment", []interface{}{}, logger)
	if err != nil {
		return
	}
	assert.NoError(t, err)
}

func Test_operationHandler_createCR(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	opsHandlerInterface := mock.NewMockoperationHandlerInterface(ctrl)
	ctx := context.Background()
	opsHandlerInterface.EXPECT().CreateCR(ctx).Return(nil, nil)
	_, err := opsHandlerInterface.CreateCR(ctx)
	if err != nil {
		return
	}
	assert.NoError(t, err)
}

func Test_operationHandler_getAndCreateOperationCR(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	opsHandlerInterface := mock.NewMockoperationHandlerInterface(ctrl)
	ctx := context.Background()
	opsHandlerInterface.EXPECT().GetAndCreateOperationCR(ctx).Return(nil, nil)
	_, err := opsHandlerInterface.GetAndCreateOperationCR(ctx)
	if err != nil {
		return
	}
	assert.NoError(t, err)
}

func Test_operationHandler_operationHandlerImpl(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	opsHandlerInterface := mock.NewMockoperationHandlerInterface(ctrl)
	ctx := context.Background()
	opsHandlerInterface.EXPECT().OperationHandlerImpl(ctx).Return(nil, nil)
	_, err := opsHandlerInterface.OperationHandlerImpl(ctx)
	if err != nil {
		return
	}
	assert.NoError(t, err)
}
