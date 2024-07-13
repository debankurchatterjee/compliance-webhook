package handler

import (
	"context"
	"encoding/json"
	"fmt"
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

func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	var admissionReview admissionv1.AdmissionReview
	ctx := context.Background()
	logger := log.From(ctx)
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
			return
		}
		admissionReview.Response, err = handleAdmissionRequest(ctx, admissionReview.Request, snowController, logger)
		fmt.Println("Admission review response ", admissionReview.Response.Result)
	}
	responseBytes, err := json.Marshal(admissionReview)
	if err != nil {
		http.Error(w, "could not marshal response", http.StatusInternalServerError)
		return
	}
	_, err = w.Write(responseBytes)
	if err != nil {
		return
	}
}
