package handler

import (
	"context"
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
	"strings"
)

// admissionOperationHandler struct to handle admission operation
type admissionOperationHandler struct {
}

// operationHandler struct to Handle only operations like create,update and delete
type operationHandler struct {
	req                    *admissionv1.AdmissionRequest
	operation              string
	currentChangeID        string
	parentChangeID         string
	createChangeID         string
	name                   string
	namespace              string
	kind                   string
	byPassPayloadInjection bool
	byPassStatusCheck      bool
	generateName           bool
	ownerReferences        []interface{}
	resource               controller.SnowResource
	logger                 logr.Logger
}

// operationHandlerInterface is an interface to Handle operation,create CR and GET CR
type operationHandlerInterface interface { // nolint
	OperationHandlerImpl(ctx context.Context) (*admissionv1.AdmissionResponse, error)
	GetAndCreateOperationCR(ctx context.Context) (*admissionv1.AdmissionResponse, error)
	CreateCR(ctx context.Context) (*admissionv1.AdmissionResponse, error)
}

// var opsHandler operationHandler

// operationHandlerFactory it is a factory interface to handle the CUD operation on the resources
type operationHandlerFactory interface {
	Handle(ctx context.Context, req *admissionv1.AdmissionRequest, operation *admissionv1.Operation, resource controller.SnowResource, name, namespace, kind string,
		ownerReferences []interface{}, logger logr.Logger) (*admissionv1.AdmissionResponse, error)
}

// Handle method with handle each operation using OperationHandlerImpl
func (a admissionOperationHandler) Handle(ctx context.Context, req *admissionv1.AdmissionRequest, operation *admissionv1.Operation, resource controller.SnowResource, name, namespace, kind string, ownerReferences []interface{}, logger logr.Logger) (*admissionv1.AdmissionResponse, error) {
	opsHandler := &operationHandler{}
	opsHandler.req = req
	opsHandler.resource = resource
	opsHandler.name = name
	opsHandler.operation = strings.ToLower(string(*operation))
	opsHandler.namespace = namespace
	opsHandler.kind = kind
	opsHandler.ownerReferences = ownerReferences
	opsHandler.logger = logger

	return opsHandler.OperationHandlerImpl(ctx)
}

// OperationHandlerImpl will Handle operations create,update and delete
func (o *operationHandler) OperationHandlerImpl(ctx context.Context) (*admissionv1.AdmissionResponse, error) {
	o.currentChangeID = k8s.GenerateChangeID(o.name, o.namespace, o.operation, o.kind)
	o.logger.Info("change id for given request", "ChangeID", o.currentChangeID)
	o.logger.Info("current resource info", "Kind", o.kind, "Name", o.name, "Namespace", o.namespace)
	o.byPassStatusCheck = true
	o.byPassPayloadInjection = false
	switch o.operation {
	case "create":
		return o.GetAndCreateOperationCR(ctx)
	case "delete":
		return o.GetAndCreateOperationCR(ctx)
	case "update":
		return o.GetAndCreateOperationCR(ctx)
	}
	return &admissionv1.AdmissionResponse{
		Allowed: false,
		Result: &metav1.Status{
			Message: "Unsupported operation",
		},
	}, nil
}

// GetAndCreateOperationCR this function will check for owner reference if the owner reference is already approved
// then it will approve the admission request else it will check if the CR is already there if not it will create the new CR and approve the request
func (o *operationHandler) GetAndCreateOperationCR(ctx context.Context) (*admissionv1.AdmissionResponse, error) {
	if len(o.ownerReferences) > 0 {
		ownRefs := k8s.ParseOwnerReference(o.ownerReferences)[0]
		o.logger.Info("found owner reference", "Name", ownRefs[1], "Kind", ownRefs[0], "Namespace", o.namespace)
		approved, err := isOwnerApproved(ctx, ownRefs[0], ownRefs[1], o.namespace, o.operation, o.resource, o.logger)
		if err != nil {
			return nil, err
		}
		if approved {
			o.logger.Info("owner reference is already approved,child resource will be default approved",
				"Name", ownRefs[1], "Kind", ownRefs[0], "Namespace", o.namespace)
			return &admissionv1.AdmissionResponse{
				Allowed: true,
				UID:     o.req.UID,
				Result: &metav1.Status{
					Code:    http.StatusOK,
					Message: "request accepted",
				},
			}, nil
		}
	}
	_, result, err := o.resource.Get(ctx, o.currentChangeID, o.req.Namespace, o.operation, o.byPassStatusCheck)
	if err != nil {
		if errors.IsNotFound(err) || !result {
			o.logger.Error(err, "unable to find service now request for the change")
			o.createChangeID = k8s.GenerateChangeID(o.req.Name, o.req.Namespace, "create", o.req.Kind.Kind)
			// dummy data to show rejection
			if o.req.Name == "nginx-app-2" && o.req.Kind.Kind == "Deployment" && o.operation == "create" {
				return &admissionv1.AdmissionResponse{
					Allowed: false,
					UID:     o.req.UID,
					Result: &metav1.Status{
						Message: "no approval request",
					},
				}, nil
			}
			o.generateName = true
		}
		return o.CreateCR(ctx)
	}
	return &admissionv1.AdmissionResponse{
		Allowed: true,
		UID:     o.req.UID,
		Result: &metav1.Status{
			Code:    http.StatusOK,
			Message: "request accepted",
		},
	}, nil
}

// CreateCR wraps around k8s dynamic client which will create the snow the CR
func (o *operationHandler) CreateCR(ctx context.Context) (*admissionv1.AdmissionResponse, error) {
	var payloadYAML = []byte{}
	// bypass payload injection for Delete operation
	if !o.byPassPayloadInjection {
		reqData := make(map[string]interface{})
		var rawRequestData []byte
		if o.req.Operation == admissionv1.Delete {
			rawRequestData = o.req.OldObject.Raw
		} else {
			rawRequestData = o.req.Object.Raw
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
	labels["snow.controller/changeID"] = o.currentChangeID
	if o.parentChangeID != "" {
		labels["snow.controller/parentChangeID"] = o.parentChangeID
	}
	if o.createChangeID != "" {
		labels["snow.controller/createChangeID"] = o.createChangeID
	}
	name, err := o.resource.Create(ctx, o.name, o.req.Namespace, o.operation, o.req.Kind.Kind, string(payloadYAML), labels, o.generateName)
	if err != nil {
		o.logger.Error(err, "unable to create the service now request for the given change")
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
		UID:     o.req.UID,
		Result: &metav1.Status{
			Code:    http.StatusOK,
			Message: fmt.Sprintf("request accepted and corresponding service now request %s has been created", name),
		},
	}, nil
}
