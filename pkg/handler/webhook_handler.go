package handler

import (
	"encoding/json"
	"github.com/compliance-webhook/internal/logutil/log"
	"github.com/compliance-webhook/pkg/controller"
	"io"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	group    = "compliance.complaince.org"
	version  = "v1"
	resource = "snows"
)

// WebhookHandler it is the http handler this will Handle the http request from the API server
func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	var admissionReview admissionv1.AdmissionReview
	ctx := r.Context()
	logger := log.From(ctx).WithName("webhook-handler")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "could not read request body", http.StatusBadRequest)
		logger.Error(err, "could not read request body", "status", http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(body, &admissionReview); err != nil {
		http.Error(w, "could not unmarshal request body", http.StatusBadRequest)
		logger.Error(err, "could not unmarshal request body", "status", http.StatusBadRequest)
		return
	}
	bodyStr := string(body)
	logger.Info("message body", "BodyStr", bodyStr)
	kind := admissionReview.Request.Kind.Kind
	patchType := admissionv1.PatchTypeJSONPatch
	// allow the Snow CR LCM operation
	if kind == "Snow" {
		admissionReview.Response = &admissionv1.AdmissionResponse{
			Allowed:   true,
			UID:       admissionReview.Request.UID,
			PatchType: &patchType,
			Result: &metav1.Status{
				Code:    http.StatusOK,
				Message: "request accepted",
			},
		}
	} else {
		snowController, err := controller.NewSnowResource(group, version, resource, true)
		if err != nil {
			logger.Error(err, "error while creating snow resource controller")
			w.Header().Set("Content-Type", "application/json")
			writer, err := w.Write([]byte(err.Error()))
			if err != nil {
				logger.Error(err, "error while parsing the error response", "WriteValue", writer)
			}
			return
		}
		admissionReview.Response, err = handleAdmissionRequest(ctx, admissionReview.Request, snowController)
		if err != nil {
			logger.Error(err, "error while handling admission request")
			w.Header().Set("Content-Type", "application/json")
			writer, err := w.Write([]byte(err.Error()))
			if err != nil {
				logger.Error(err, "error while parsing the error response", "WriteValue", writer)
			}
			return
		}
	}
	responseBytes, err := json.Marshal(admissionReview)
	if err != nil {
		http.Error(w, "could not marshal response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(responseBytes)
	if err != nil {
		return
	}
}
