package handler

import (
	"context"
	"github.com/compliance-webhook/internal/logutil/log"
	"github.com/compliance-webhook/pkg/controller"
	"github.com/go-logr/logr"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
)

var admissionReview = `
{
      "kind": "ReplicaSet",
      "apiVersion": "apps/v1",
      "metadata": {
        "name": "nginx-app-cbdccf466",
        "namespace": "test",
        "creationTimestamp": null,
        "labels": {
          "app": "nginx",
          "pod-template-hash": "cbdccf466"
        },
        "annotations": {
          "deployment.kubernetes.io/desired-replicas": "1",
          "deployment.kubernetes.io/max-replicas": "2",
          "deployment.kubernetes.io/revision": "1"
        },
        "ownerReferences": [
          {
            "apiVersion": "apps/v1",
            "kind": "Deployment",
            "name": "nginx-app",
            "uid": "fe768b9d-011f-46a0-ac74-6b0256c6acf6",
            "controller": true,
            "blockOwnerDeletion": true
          }
        ],
        "managedFields": [
          {
            "manager": "kube-controller-manager",
            "operation": "Update",
            "apiVersion": "apps/v1",
            "time": "2024-07-14T20:25:52Z",
            "fieldsType": "FieldsV1",
            "fieldsV1": {
              "f:metadata": {
                "f:annotations": {
                  ".": {},
                  "f:deployment.kubernetes.io/desired-replicas": {},
                  "f:deployment.kubernetes.io/max-replicas": {},
                  "f:deployment.kubernetes.io/revision": {}
                },
                "f:labels": {
                  ".": {},
                  "f:app": {},
                  "f:pod-template-hash": {}
                },
                "f:ownerReferences": {
                  ".": {},
                  "k:{\"uid\":\"fe768b9d-011f-46a0-ac74-6b0256c6acf6\"}": {}
                }
              },
              "f:spec": {
                "f:replicas": {},
                "f:selector": {},
                "f:template": {
                  "f:metadata": {
                    "f:labels": {
                      ".": {},
                      "f:app": {},
                      "f:pod-template-hash": {}
                    }
                  },
                  "f:spec": {
                    "f:containers": {
                      "k:{\"name\":\"nginx\"}": {
                        ".": {},
                        "f:image": {},
                        "f:imagePullPolicy": {},
                        "f:name": {},
                        "f:ports": {
                          ".": {},
                          "k:{\"containerPort\":80,\"protocol\":\"TCP\"}": {
                            ".": {},
                            "f:containerPort": {},
                            "f:protocol": {}
                          }
                        },
                        "f:resources": {},
                        "f:terminationMessagePath": {},
                        "f:terminationMessagePolicy": {}
                      }
                    },
                    "f:dnsPolicy": {},
                    "f:restartPolicy": {},
                    "f:schedulerName": {},
                    "f:securityContext": {},
                    "f:terminationGracePeriodSeconds": {}
                  }
                }
              }
            }
          }
        ]
      },
      "spec": {
        "replicas": 1,
        "selector": {
          "matchLabels": {
            "app": "nginx",
            "pod-template-hash": "cbdccf466"
          }
        },
        "template": {
          "metadata": {
            "creationTimestamp": null,
            "labels": {
              "app": "nginx",
              "pod-template-hash": "cbdccf466"
            }
          },
          "spec": {
            "containers": [
              {
                "name": "nginx",
                "image": "nginx:1.14.2",
                "ports": [
                  {
                    "containerPort": 80,
                    "protocol": "TCP"
                  }
                ],
                "resources": {},
                "terminationMessagePath": "/dev/termination-log",
                "terminationMessagePolicy": "File",
                "imagePullPolicy": "IfNotPresent"
              }
            ],
            "restartPolicy": "Always",
            "terminationGracePeriodSeconds": 30,
            "dnsPolicy": "ClusterFirst",
            "securityContext": {},
            "schedulerName": "default-scheduler"
          }
        }
      },
      "status": {
        "replicas": 0
      }
    }`

func parseAdmissionReviewRequest() *admissionv1.AdmissionRequest {
	reviewRequest := &admissionv1.AdmissionRequest{}
	reviewRequest.Object.Raw = []byte(admissionReview)
	return reviewRequest
}

func Test_handleAdmissionRequest(t *testing.T) {
	type args struct {
		ctx      context.Context
		req      *admissionv1.AdmissionRequest
		resource controller.SnowResource
	}
	snowController, err := controller.NewSnowResource(group, version, resource, false)
	if err != nil {
		return
	}
	tests := []struct {
		name    string
		args    args
		want    *admissionv1.AdmissionResponse
		wantErr bool
	}{
		{name: "t1", args: struct {
			ctx      context.Context
			req      *admissionv1.AdmissionRequest
			resource controller.SnowResource
		}{ctx: context.Background(), req: parseAdmissionReviewRequest(), resource: snowController}, want: &admissionv1.AdmissionResponse{
			Allowed: false,
			UID:     "",
			Result: &metav1.Status{
				Message: "Unsupported operation",
			},
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := handleAdmissionRequest(tt.args.ctx, tt.args.req, tt.args.resource)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleAdmissionRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleAdmissionRequest() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isOwnerApproved(t *testing.T) {
	type args struct {
		ctx       context.Context
		kind      string
		name      string
		namespace string
		operation string
		resource  controller.SnowResource
		logger    logr.Logger
	}
	snowController, err := controller.NewSnowResource(group, version, resource, false)
	if err != nil {
		return
	}
	logger := log.From(context.Background())
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{name: "t1", args: struct {
			ctx       context.Context
			kind      string
			name      string
			namespace string
			operation string
			resource  controller.SnowResource
			logger    logr.Logger
		}{ctx: context.Background(), kind: "Deployment", name: "test-nginx-app", namespace: "default", operation: "create", resource: snowController, logger: logger}, want: false, wantErr: true},
		{name: "t2", args: struct {
			ctx       context.Context
			kind      string
			name      string
			namespace string
			operation string
			resource  controller.SnowResource
			logger    logr.Logger
		}{ctx: context.Background(), kind: "Deployment", name: "nginx-app", namespace: "default", operation: "create", resource: snowController, logger: logger}, want: true, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isOwnerApproved(tt.args.ctx, tt.args.kind, tt.args.name, tt.args.namespace, tt.args.operation, tt.args.resource, tt.args.logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("isOwnerApproved() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("isOwnerApproved() got = %v, want %v", got, tt.want)
			}
		})
	}
}
