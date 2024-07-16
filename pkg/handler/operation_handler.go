package handler

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/compliance-webhook/internal/k8s"
	"github.com/compliance-webhook/pkg/controller"
	"github.com/go-logr/logr"
	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"sigs.k8s.io/yaml"
)

// operationHandler will handle the create request
func operationHandler(ctx context.Context,
	req *admissionv1.AdmissionRequest,
	resource controller.SnowResource,
	name, operation, namespace, kind string,
	ownerReferences []interface{}, logger logr.Logger) (*admissionv1.AdmissionResponse, error) {

	changeStr := fmt.Sprintf("%s-%s-%s-%s", name, operation, namespace, kind)
	changeID := md5.Sum([]byte(changeStr))
	changeIDStr := hex.EncodeToString(changeID[:])
	logger.Info("change id for given request", "ChangeID", changeIDStr)
	logger.Info("current resource info", "Kind", kind, "Name", name, "Namespace", namespace)
	if operation == "create" {
		if len(ownerReferences) > 0 {
			ownRefs := k8s.ParseOwnerReference(ownerReferences)[0]
			logger.Info("found owner reference",
				"Name", ownRefs[1],
				"Kind", ownRefs[0],
				"Namespace", namespace)
			approved, err := isOwnerApproved(ctx, ownRefs[0], ownRefs[1], namespace, operation, resource, logger)
			if err != nil {
				return nil, err
			}
			if approved {
				logger.Info("owner reference is already approved,child resource will be default approved",
					"Name", ownRefs[1],
					"Kind", ownRefs[0],
					"Namespace", namespace)
				return &admissionv1.AdmissionResponse{
					Allowed: true,
					UID:     req.UID,
					Result: &metav1.Status{
						Code:    http.StatusOK,
						Message: "request accepted",
					},
				}, nil
			}
		}
	}
	result, err := resource.Get(ctx, changeIDStr, req.Namespace, operation)
	if err != nil {
		if errors.IsNotFound(err) || !result {
			logger.Error(err, "unable to find service now request for the change")
			reqData := make(map[string]interface{})
			err := json.Unmarshal(req.Object.Raw, &reqData)
			if err != nil {
				return nil, err
			}
			rawRequestData := req.Object.Raw
			metadata, ok := reqData["metadata"].(map[string]interface{})
			if ok {
				annotations, ok := metadata["annotations"].(map[string]interface{})
				if ok {
					appliedConfig, ok := annotations["kubectl.kubernetes.io/last-applied-configuration"].(string)
					if ok {
						rawRequestData = []byte(appliedConfig)
					}
				}
			}
			payloadYAML, err := yaml.JSONToYAML(rawRequestData)
			if err != nil {
				return nil, err
			}
			err = resource.Create(ctx, req.Name, req.Namespace, operation, req.Kind.Kind, string(payloadYAML))
			if err != nil {
				logger.Error(err, "unable to create the service now request for the given change")
				return &admissionv1.AdmissionResponse{
					Allowed: false,
					Result: &metav1.Status{
						Message: fmt.Sprintf("unable to create the snow request for the given CR,"+
							"please try creating manually %v", err),
					},
				}, nil
			} else {
				return &admissionv1.AdmissionResponse{
					Allowed: true,
					UID:     req.UID,
					Result: &metav1.Status{
						Code:    http.StatusOK,
						Message: "request accepted and corresponding service now request has been created",
					},
				}, nil
			}
		}
	}
	return &admissionv1.AdmissionResponse{
		Allowed: true,
		UID:     req.UID,
		Result: &metav1.Status{
			Code:    http.StatusOK,
			Message: "request accepted",
		},
	}, nil
}
