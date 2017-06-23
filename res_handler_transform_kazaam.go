package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/mitchellh/mapstructure"
)

type ResponseTransformKazaamOptions struct {
}

type ResponseTransformKazaamMiddleware struct {
	Spec   *APISpec
	config ResponseTransformKazaamOptions
}

func (rtk ResponseTransformKazaamMiddleware) New(c interface{}, spec *APISpec) (TykResponseHandler, error) {
	handler := ResponseTransformKazaamMiddleware{Spec: spec}

	if err := mapstructure.Decode(c, &handler.config); err != nil {
		log.Error(err)
		return nil, err
	}

	log.Debug("Response Transform Kazaam processor initialised")

	return handler, nil
}

func (rtk ResponseTransformKazaamMiddleware) HandleResponse(rw http.ResponseWriter, res *http.Response, req *http.Request, ses *SessionState) error {
	// TODO: This should only target specific paths
	_, versionPaths, _, _ := rtk.Spec.GetVersionData(req)
	found, meta := rtk.Spec.CheckSpecMatchesStatus(req.URL.Path, req.Method, versionPaths, TransformedKazaamResponse)
	if !found {
		return nil
	}
	tmeta := meta.(*TransformKazaamSpec)

	log.Debug("[KAZAAMTRANSFORMRESPONSE] We are in HandleResponse")

	// Read the body:
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var transformed []byte
	if transformed, err = tmeta.Spec.TransformInPlace(body); err != nil {
		log.WithFields(logrus.Fields{
			"prefix":      "outbound-transform-kazaam",
			"server_name": rtk.Spec.Proxy.TargetURL,
			"api_id":      rtk.Spec.APIID,
			"path":        req.URL.Path,
		}).Error("Failed to apply Kaazam transformation to request: ", err)
	}

	bodyBuffer := bytes.NewBuffer(transformed)

	res.Body = ioutil.NopCloser(bodyBuffer)
	res.ContentLength = int64(bodyBuffer.Len())
	res.Header.Set("Content-Length", strconv.Itoa(bodyBuffer.Len()))

	return nil
}
