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
)

var OperationHandlerFactory operationHandlerFactory = &admissionOperationHandler{}

func handleAdmissionRequest(ctx context.Context,
	req *admissionv1.AdmissionRequest,
	resource controller.SnowResource) (*admissionv1.AdmissionResponse, error) {
	logger := log.From(ctx).WithName("webhook-admission-handler")
	logger.Info("handling operation", "Operation", req.Operation, "Kind", req.Kind.Kind)
	name := req.Name
	kind := req.Kind.Kind
	namespace := req.Namespace
	// Verify the object has owner reference for e.g. like ReplicaSet or StatefulSets can have owner ref to Deployment
	ownerReferences, err := k8s.FindOwnerReferenceFromRawObject(req)
	if err != nil {
		return nil, err
	}
	logger.Info("owner reference of the request", "OwnerReference", ownerReferences)
	logger.Info("handling operation", "Operation",
		req.Operation, "Kind", req.Kind.Kind,
		"ChangeName", name,
		"Namespace", namespace)

	return OperationHandlerFactory.handle(ctx, req, &req.Operation, resource, name, namespace, kind, ownerReferences, logger)
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
	logger.Info("change for owner request", "Name", name, "Operation", operation, "Namespace", namespace, "Kind", kind)
	logger.Info("change id for the parent request", "ChangeID", OwnerChangeIDStr)
	return resource.Get(ctx, OwnerChangeIDStr, "", operation, true)
}
