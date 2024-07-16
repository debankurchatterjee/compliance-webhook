package handler

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/compliance-webhook/internal/k8s"
	"github.com/compliance-webhook/internal/logutil/log"
	"github.com/compliance-webhook/pkg/controller"
	"github.com/go-logr/logr"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func handleAdmissionRequest(ctx context.Context,
	req *admissionv1.AdmissionRequest,
	resource controller.SnowResource) (*admissionv1.AdmissionResponse, error) {
	logger := log.From(ctx).WithName("webhook-admission-handler")
	logger.Info("handling operation", "Operation", req.Operation, "Kind", req.Kind.Kind)
	name := req.Name
	kind := req.Kind.Kind
	namespace := req.Namespace
	// Verify the object has owner reference for e.g. like ReplicaSet or StatefulSets can have owner ref to Deployment
	ownerReferences, err := k8s.FindOwnerReferenceFromRawObject(req.Object.Raw)
	if err != nil {
		return nil, err
	}
	logger.Info("handling operation", "Operation",
		req.Operation, "Kind", req.Kind.Kind,
		"ChangeName", name,
		"Namespace", namespace)
	switch req.Operation {
	case admissionv1.Create:
		return operationHandler(ctx, req, resource, name, "create", namespace, kind, ownerReferences, logger)
	case admissionv1.Update:
		return operationHandler(ctx, req, resource, name, "update", namespace, kind, ownerReferences, logger)
	default:
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: "Unsupported operation",
			},
		}, nil
	}
	return nil, nil
}
func isOwnerApproved(ctx context.Context,
	kind,
	name,
	namespace,
	operation string,
	resource controller.SnowResource, logger logr.Logger) (bool, error) {
	changeStr := fmt.Sprintf("%s-%s-%s-%s", name, operation, namespace, kind)
	changeID := md5.Sum([]byte(changeStr))
	OwnerChangeIDStr := hex.EncodeToString(changeID[:])
	logger.Info("change id for the parent request", "ChangeID", changeStr)
	return resource.Get(ctx, OwnerChangeIDStr, "", operation)
}
