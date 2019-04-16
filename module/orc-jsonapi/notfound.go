package jsonapi

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

func EndpointNotFoundHandler(w http.ResponseWriter, req *http.Request) {
	fields := logrus.Fields{
		"method": req.Method,
		"path":   req.URL.Path,
	}
	logrus.WithFields(fields).Warningf("API endpoint not found")
	writeResponse(w, http.StatusNotFound, nil)
}
