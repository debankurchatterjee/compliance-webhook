package handler

import (
	"context"
	"crypto/md5" //nolint
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
	"strconv"
	"strings"
)

// admissionOperationHandler struct to handle admission operation
type admissionOperationHandler struct {
}

// operationHandler struct to Handle only operations like create,update and delete
type operationHandler struct{}

// operationHandlerInterface is an interface to Handle operation,create CR and GET CR
type operationHandlerInterface interface {
	OperationHandlerImpl(ctx context.Context,
		req *admissionv1.AdmissionRequest,
		resource controller.SnowResource,
		name, operation, namespace, kind string,
		ownerReferences []interface{}, logger logr.Logger) (*admissionv1.AdmissionResponse, error)
	GetAndCreateOperationCR(ctx context.Context, req *admissionv1.AdmissionRequest, operation, changeIDStr, namespace string, byPassStatusCheck, byPassPayloadInjection bool, logger logr.Logger, resource controller.SnowResource, ownerReferences []interface{}) (*admissionv1.AdmissionResponse, error)
	CreateCR(ctx context.Context, req *admissionv1.AdmissionRequest, operation, changeIDStr, parentChangeID, name string, byPassPayloadInjection bool, logger logr.Logger, resource controller.SnowResource, generateName bool) (*admissionv1.AdmissionResponse, error)
}

var opsHandler operationHandler

// operationHandlerFactory it is a factory interface to handle the CUD operation on the resources
type operationHandlerFactory interface {
	Handle(ctx context.Context, req *admissionv1.AdmissionRequest, operation *admissionv1.Operation, resource controller.SnowResource, name, namespace, kind string,
		ownerReferences []interface{}, logger logr.Logger) (*admissionv1.AdmissionResponse, error)
}

// Handle method with handle each operation using OperationHandlerImpl
func (a admissionOperationHandler) Handle(ctx context.Context, req *admissionv1.AdmissionRequest, operation *admissionv1.Operation, resource controller.SnowResource, name, namespace, kind string, ownerReferences []interface{}, logger logr.Logger) (*admissionv1.AdmissionResponse, error) {
	return opsHandler.OperationHandlerImpl(ctx, req, resource, name, strings.ToLower(string(*operation)), namespace, kind, ownerReferences, logger)
}

// operationHandlerImpl will Handle operations create,update and delete
func (o *operationHandler) OperationHandlerImpl(ctx context.Context,
	req *admissionv1.AdmissionRequest,
	resource controller.SnowResource,
	name, operation, namespace, kind string,
	ownerReferences []interface{}, logger logr.Logger) (*admissionv1.AdmissionResponse, error) {

	changeStr := fmt.Sprintf("%s-%s-%s-%s", name, operation, namespace, kind)
	changeID := md5.Sum([]byte(changeStr)) // nolint
	changeIDStr := hex.EncodeToString(changeID[:])
	logger.Info("change id for given request", "ChangeID", changeIDStr)
	logger.Info("current resource info", "Kind", kind, "Name", name, "Namespace", namespace)
	switch operation {
	case "create":
		return opsHandler.GetAndCreateOperationCR(ctx, req, "create", changeIDStr, namespace, true, false, logger, resource, ownerReferences)
	case "delete":
		return opsHandler.GetAndCreateOperationCR(ctx, req, "delete", changeIDStr, namespace, true, false, logger, resource, ownerReferences)
	case "update":
		return opsHandler.GetAndCreateOperationCR(ctx, req, "update", changeIDStr, namespace, true, false, logger, resource, ownerReferences)
	}
	return &admissionv1.AdmissionResponse{
		Allowed: false,
		Result: &metav1.Status{
			Message: "Unsupported operation",
		},
	}, nil
}

// getAndCreateOperationCR this function will check for owner reference if the owner reference is already approved
// then it will approve the admission request else it will check if the CR is already there if not it will create the new CR and approve the request
func (o *operationHandler) GetAndCreateOperationCR(ctx context.Context, req *admissionv1.AdmissionRequest, operation, changeIDStr, namespace string, byPassStatusCheck, byPassPayloadInjection bool, logger logr.Logger, resource controller.SnowResource, ownerReferences []interface{}) (*admissionv1.AdmissionResponse, error) {
	if len(ownerReferences) > 0 {
		ownRefs := k8s.ParseOwnerReference(ownerReferences)[0]
		logger.Info("found owner reference", "Name", ownRefs[1], "Kind", ownRefs[0], "Namespace", namespace)
		approved, err := isOwnerApproved(ctx, ownRefs[0], ownRefs[1], namespace, operation, resource, logger)
		if err != nil {
			return nil, err
		}
		if approved {
			logger.Info("owner reference is already approved,child resource will be default approved",
				"Name", ownRefs[1], "Kind", ownRefs[0], "Namespace", namespace)
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
	name, result, err := resource.Get(ctx, changeIDStr, req.Namespace, operation, byPassStatusCheck)
	if err != nil {
		if errors.IsNotFound(err) || !result {
			logger.Error(err, "unable to find service now request for the change")
			parentChangeID := ""
			if operation == "update" {
				parentChangeID = changeIDStr
			}
			return opsHandler.CreateCR(ctx, req, operation, changeIDStr, parentChangeID, req.Name, byPassPayloadInjection, logger, resource, true)
		}
	}
	if req.Operation == admissionv1.Update && name != "" {
		logger.Info("current name", "Name", name)
		res := strings.Split(name, "-")
		revision := res[len(res)-1]
		revisionNum, err := strconv.Atoi(revision)
		if err != nil {
			name = fmt.Sprintf("%s-%d", name, 1)
			logger.Info("current name with no revision", "Name", name)
		} else {
			revisionNum++
			res[len(res)-1] = fmt.Sprintf("%d", revisionNum)
			name = strings.Join(res, "-")
			logger.Info("current name with updated revision", "Name", name)
		}
		parentChangeID := changeIDStr
		changeID := md5.Sum([]byte(name)) //nolint
		changeIDStr = hex.EncodeToString(changeID[:])
		logger.Info("change id for given request", "ChangeID", changeIDStr)
		return opsHandler.CreateCR(ctx, req, operation, changeIDStr, parentChangeID, name, byPassPayloadInjection, logger, resource, false)
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

// createCR wraps around k8s dynamic client which will create the snow the CR
func (o *operationHandler) CreateCR(ctx context.Context, req *admissionv1.AdmissionRequest, operation, changeIDStr, parentChangeID, name string, byPassPayloadInjection bool, logger logr.Logger, resource controller.SnowResource, generateName bool) (*admissionv1.AdmissionResponse, error) {
	var payloadYAML = []byte{}
	// bypass payload injection for Delete operation
	if !byPassPayloadInjection {
		reqData := make(map[string]interface{})
		var rawRequestData []byte
		if req.Operation == admissionv1.Delete {
			rawRequestData = req.OldObject.Raw
		} else {
			rawRequestData = req.Object.Raw
		}
		err := json.Unmarshal(rawRequestData, &reqData)
		if err != nil {
			return nil, err
		}
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
		payloadYAML, err = yaml.JSONToYAML(rawRequestData)
		if err != nil {
			return nil, err
		}
	}
	labels := make(map[string]string)
	labels["snow.controller/changeID"] = changeIDStr
	if parentChangeID != "" {
		labels["snow.controller/parentChangeID"] = parentChangeID
	}
	name, err := resource.Create(ctx, name, req.Namespace, operation, req.Kind.Kind, string(payloadYAML), labels, generateName)
	if err != nil {
		logger.Error(err, "unable to create the service now request for the given change")
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: fmt.Sprintf("unable to create the snow request for the given CR,"+
					"please try creating manually %v", err),
			},
		}, nil
	}
	return &admissionv1.AdmissionResponse{
		Allowed: true,
		UID:     req.UID,
		Result: &metav1.Status{
			Code:    http.StatusOK,
			Message: fmt.Sprintf("request accepted and corresponding service now request %s has been created", name),
		},
	}, nil
}
