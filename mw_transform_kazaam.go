package main

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
)

//func WrappedCharsetReader(s string, i io.Reader) (io.Reader, error) {
//	return charset.NewReader(i, s)
//}

// TransformMiddleware is a middleware that will apply a template to a request body to transform it's contents ready for an upstream API
type TransformKazaamMiddleware struct {
	*TykMiddleware
}

func (t *TransformKazaamMiddleware) GetName() string {
	return "TransformKazaamMiddleware"
}

func (t *TransformKazaamMiddleware) IsEnabledForSpec() bool {
	for _, version := range t.Spec.VersionData.Versions {
		if len(version.ExtendedPaths.TransformKazaam) > 0 {
			return true
		}
	}
	return false
}

// ProcessRequest will run any checks on the request on the way through the system, return an error to have the chain fail
func (t *TransformKazaamMiddleware) ProcessRequest(w http.ResponseWriter, r *http.Request, _ interface{}) (error, int) {
	_, versionPaths, _, _ := t.Spec.GetVersionData(r)
	found, meta := t.Spec.CheckSpecMatchesStatus(r.URL.Path, r.Method, versionPaths, TransformedKazaam)
	if !found {
		return nil, 200
	}
	tmeta := meta.(*TransformKazaamSpec)

	// Read the body:
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, http.StatusBadRequest
	}

	// Apply to template
	var transformed []byte
	if transformed, err = tmeta.Spec.TransformInPlace(body); err != nil {
		log.WithFields(logrus.Fields{
			"prefix":      "inbound-transform-kazaam",
			"server_name": t.Spec.Proxy.TargetURL,
			"api_id":      t.Spec.APIID,
			"path":        r.URL.Path,
		}).Error("Failed to apply Kaazam transformation to request: ", err)
	}

	bodyBuffer := bytes.NewBuffer(transformed)
	r.Body = ioutil.NopCloser(bodyBuffer)
	r.ContentLength = int64(bodyBuffer.Len())

	return nil, 200
}
