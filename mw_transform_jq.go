package main

import (
	"bytes"
	//	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
)

type TransformJQMiddleware struct {
	*BaseMiddleware
}

func (t *TransformJQMiddleware) GetName() string {
	return "TransformJQMiddleware"
}

func (t *TransformJQMiddleware) IsEnabledForSpec() bool {
	for _, version := range t.Spec.VersionData.Versions {
		if len(version.ExtendedPaths.TransformJQ) > 0 {
			return true
		}
	}
	return false
}

// ProcessRequest will run any checks on the request on the way through the system, return an error to have the chain fail
func (t *TransformJQMiddleware) ProcessRequest(w http.ResponseWriter, r *http.Request, _ interface{}) (error, int) {
	_, versionPaths, _, _ := t.Spec.GetVersionData(r)
	found, meta := t.Spec.CheckSpecMatchesStatus(r, versionPaths, TransformedJQ)
	if !found {
		return nil, 200
	}
	err := transformJQBody(r, meta.(*TransformJQSpec), t.Spec.EnableContextVars)
	if err != nil {
		log.WithFields(logrus.Fields{
			"prefix":      "inbound-transform-jq",
			"server_name": t.Spec.Proxy.TargetURL,
			"api_id":      t.Spec.APIID,
			"path":        r.URL.Path,
		}).Error(err)
		return err, 415
	}
	return nil, 200
}

func transformJQBody(r *http.Request, t *TransformJQSpec, contextVars bool) error {
	// Read the body:
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	// XXX
	// if contextVars {
	//	bodyData["_tyk_context"] = ctxGetData(r)
	//}

	err = t.JQFilter.HandleJson(string(body))
	if err != nil {
		return errors.New("Input is not a valid JSON")
	}

	filterResult := t.JQFilter.Next()
	if filterResult {
		transformed := t.JQFilter.ValueJson()
		bodyBuffer := bytes.NewBufferString(transformed)
		r.Body = ioutil.NopCloser(bodyBuffer)
		r.ContentLength = int64(bodyBuffer.Len())
	} else {
		return errors.New("Errors while applying JQ filter to input")
	}

	return nil
}
