package handler

import (
	"encoding/json"
	"io"
	admissionv1 "k8s.io/api/admission/v1"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestWebhookHandler(t *testing.T) {
	reqData, err := os.Open("/Users/cdebankur/go/src/github.com/compliance-webhook/resources/create-resource.json")
	if err != nil {
		return
	}
	req, err := http.NewRequest("POST", "/mutate", reqData)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}
	rec := httptest.NewRecorder()
	WebhookHandler(rec, req)

	res := rec.Result()
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(res.Body)

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %v", res.Status)
	}

	contentType := res.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected content-type application/json; got %s", contentType)
	}

	var response *admissionv1.AdmissionResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
}
