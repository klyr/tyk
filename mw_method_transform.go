package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/TykTechnologies/tyk/apidef"
)

// TransformMiddleware is a middleware that will apply a template to a request body to transform it's contents ready for an upstream API
type TransformMethod struct {
	*TykMiddleware
}

func (t *TransformMethod) GetName() string {
	return "TransformMethod"
}

func (t *TransformMethod) IsEnabledForSpec() bool {
	for _, version := range t.Spec.VersionData.Versions {
		if len(version.ExtendedPaths.MethodTransforms) > 0 {
			return true
		}
	}
	return false
}

// ProcessRequest will run any checks on the request on the way through the system, return an error to have the chain fail
func (t *TransformMethod) ProcessRequest(w http.ResponseWriter, r *http.Request, _ interface{}) (error, int) {
	_, versionPaths, _, _ := t.Spec.GetVersionData(r)
	found, meta := t.Spec.CheckSpecMatchesStatus(r.URL.Path, r.Method, versionPaths, MethodTransformed)
	if found {
		mmeta := meta.(*apidef.MethodTransformMeta)

		switch strings.ToUpper(mmeta.ToMethod) {
		case "GET":
			r.Method = "GET"
			return nil, 200
		case "POST":
			r.Method = "POST"
			return nil, 200
		case "PUT":
			r.Method = "PUT"
			return nil, 200
		case "DELETE":
			r.Method = "DELETE"
			return nil, 200
		case "OPTIONS":
			r.Method = "OPTIONS"
			return nil, 200
		case "PATCH":
			r.Method = "PATCH"
			return nil, 200
		default:
			return errors.New("Method not allowed"), 405
		}

	}

	return nil, 200
}
