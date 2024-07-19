package handler

import (
	"context"
	"github.com/compliance-webhook/internal/logutil/log"
	"github.com/compliance-webhook/pkg/controller"
	"github.com/go-logr/logr"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"reflect"
	"testing"
)

func Test_admissionOperationHandler_handle(t *testing.T) {
	type args struct {
		ctx             context.Context
		req             *admissionv1.AdmissionRequest
		operation       *admissionv1.Operation
		resource        controller.SnowResource
		name            string
		namespace       string
		kind            string
		ownerReferences []interface{}
		logger          logr.Logger
	}

	snowController, err := controller.NewSnowResource(group, version, resource, false)
	if err != nil {
		return
	}
	ops := admissionv1.Update
	logger := log.From(context.Background())
	tests := []struct {
		name    string
		args    args
		want    *admissionv1.AdmissionResponse
		wantErr bool
	}{
		{name: "t1", args: struct {
			ctx             context.Context
			req             *admissionv1.AdmissionRequest
			operation       *admissionv1.Operation
			resource        controller.SnowResource
			name            string
			namespace       string
			kind            string
			ownerReferences []interface{}
			logger          logr.Logger
		}{ctx: context.Background(), req: parseAdmissionReviewRequest(), operation: &ops, resource: snowController, name: "nginx-app", namespace: "default", kind: "Deployment", ownerReferences: []interface{}{}, logger: logger}, want: &admissionv1.AdmissionResponse{
			Allowed: true,
			UID:     "",
			Result: &metav1.Status{
				Code:    http.StatusOK,
				Message: "request accepted",
			}}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := admissionOperationHandler{}
			got, err := a.handle(tt.args.ctx, tt.args.req, tt.args.operation, tt.args.resource, tt.args.name, tt.args.namespace, tt.args.kind, tt.args.ownerReferences, tt.args.logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("handle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handle() got = %v, want %v", got, tt.want)
			}
		})
	}
}
