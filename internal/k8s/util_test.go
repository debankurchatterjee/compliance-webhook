package k8s

import (
	admissionv1 "k8s.io/api/admission/v1"
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

func getOwnerRef() []interface{} {
	req := parseAdmissionReviewRequest()
	object, err := FindOwnerReferenceFromRawObject(req)
	if err != nil {
		return nil
	}
	return object
}

func TestFindOwnerReferenceFromRawObject(t *testing.T) {
	type args struct {
		req *admissionv1.AdmissionRequest
	}
	tests := []struct {
		name    string
		args    args
		want    []interface{}
		wantErr bool
	}{
		{name: "t1", args: struct{ req *admissionv1.AdmissionRequest }{req: parseAdmissionReviewRequest()}, want: nil, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := FindOwnerReferenceFromRawObject(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindOwnerReferenceFromRawObject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestParseOwnerReference(t *testing.T) {
	type args struct {
		refs []interface{}
	}
	tests := []struct {
		name string
		args args
		want [][2]string
	}{
		{name: "t1", args: struct{ refs []interface{} }{refs: getOwnerRef()}, want: [][2]string{{"Deployment", "nginx-app"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseOwnerReference(tt.args.refs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseOwnerReference() = %v, want %v", got, tt.want)
			}
		})
	}
}
