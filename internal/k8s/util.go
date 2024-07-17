package k8s

import (
	"encoding/json"
	"fmt"
	admissionv1 "k8s.io/api/admission/v1"
)

func FindOwnerReferenceFromRawObject(req *admissionv1.AdmissionRequest) ([]interface{}, error) {
	if req == nil {
		return []interface{}{}, nil
	}
	var obj []byte
	if req.Operation == admissionv1.Delete {
		obj = req.OldObject.Raw
	} else {
		obj = req.Object.Raw
	}
	rawObj := make(map[string]interface{})
	if err := json.Unmarshal(obj, &rawObj); err != nil {
		return nil, fmt.Errorf("could not unmarshal raw object: %v", err)
	}
	metadata, ok := rawObj["metadata"].(map[string]interface{})
	if !ok {
		return nil, nil
	}
	ownerReferences, ok := metadata["ownerReferences"].([]interface{})
	if !ok || len(ownerReferences) == 0 {
		return nil, nil
	}
	return ownerReferences, nil
}

func ParseOwnerReference(refs []interface{}) [][2]string {
	results := make([][2]string, 0)
	for _, ref := range refs {
		refMap, ok := ref.(map[string]interface{})
		if !ok {
			continue
		}
		kind, name := "", ""
		if kind, ok = refMap["kind"].(string); !ok {
			continue
		}
		if name, ok = refMap["name"].(string); !ok {
			continue
		}
		results = append(results, [2]string{kind, name})
	}
	return results
}
