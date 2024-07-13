package handler

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/compliance-webhook/pkg/controller"
	"github.com/go-logr/logr"
	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
)

func handleAdmissionRequest(ctx context.Context,
	req *admissionv1.AdmissionRequest,
	resource controller.SnowResource, logger logr.Logger) (*admissionv1.AdmissionResponse, error) {
	logger.Info("handling operation", "Operation", req.Operation, "Kind", req.Kind.Kind)
	name := req.Name
	kind := req.Kind.Kind
	namespace := req.Namespace
	switch req.Operation {
	case admissionv1.Create:
		operation := "create"
		changeStr := fmt.Sprintf("%s-%s-%s-%s", name, operation, namespace, kind)
		changeID := md5.Sum([]byte(changeStr))
		changeIDStr := hex.EncodeToString(changeID[:])
		result, err := resource.Get(ctx, changeIDStr, req.Namespace, "create")
		if err != nil {
			if errors.IsNotFound(err) {
				logger.Error(err, "unable to find service now request for the change")
				return &admissionv1.AdmissionResponse{
					Allowed: false,
					UID:     req.UID,
					Result: &metav1.Status{
						Code:    http.StatusForbidden,
						Message: "approved service now request not found for the resource,please raise a request",
					},
				}, nil
			} else {
				// TODO create a new Snow CR
			}
		}
		return &admissionv1.AdmissionResponse{
			Allowed: result,
			UID:     req.UID,
			Result: &metav1.Status{
				Code:    http.StatusOK,
				Message: "request accepted",
			},
		}, nil
	case admissionv1.Update:
		// Add logic to handle update operation
		return &admissionv1.AdmissionResponse{Allowed: true}, nil
	case admissionv1.Delete:
		// Add logic to handle delete operation
		return &admissionv1.AdmissionResponse{Allowed: true}, nil
	case admissionv1.Connect:
		// Add logic to handle connect operation
		return &admissionv1.AdmissionResponse{Allowed: true}, nil
	default:
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: "Unsupported operation",
			},
		}, nil
	}
}
