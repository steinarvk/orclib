package jsonapi

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
)

type genericHandler func(http.ResponseWriter, *http.Request) (interface{}, error)

type EndpointWrapper func(genericHandler) genericHandler

func applyWrappers(h genericHandler, wrappers []EndpointWrapper) genericHandler {
	for _, w := range wrappers {
		h = w(h)
	}
	return h
}

func writeResponse(w http.ResponseWriter, code int, data interface{}) error {
	if data == nil {
		resp := BasicResponse{
			Ok: code >= 200 && code <= 299,
		}
		if !resp.Ok {
			resp.Error = http.StatusText(code)
		}
		data = &resp
	}
	w.Header()["Content-Type"] = []string{"application/json"}
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func wrapForErrorWriting(next genericHandler) Handler {
	return func(w http.ResponseWriter, req *http.Request) error {
		resp, err := next(w, req)
		if err != nil {
			code := getErrorCode(err)
			if resp == nil {
				resp = &BasicResponse{
					Ok:    false,
					Error: getErrorMessage(err),
				}
			}

			if responseWriteErr := writeResponse(w, code, resp); responseWriteErr != nil {
				logrus.Warningf("error writing response: %v", responseWriteErr)
			}
		}
		// If err == nil we assume the next already wrote the response.

		// Continue to return error for tracking.
		return err
	}
}

func wrapForResponseAndErrorWriting(next genericHandler) Handler {
	return wrapForErrorWriting(func(w http.ResponseWriter, req *http.Request) (interface{}, error) {
		resp, err := next(w, req)
		if err != nil {
			// error wrapper will take care of adjusting the response.
			return resp, err
		}

		if resp == nil {
			resp = &BasicResponse{
				Ok: true,
			}
		}

		if responseWriteErr := writeResponse(w, http.StatusOK, resp); responseWriteErr != nil {
			logrus.Warningf("error writing response: %v", responseWriteErr)
		}

		return nil, nil
	})
}

func WrapOnlyErrors(next func(http.ResponseWriter, *http.Request) error, more ...EndpointWrapper) Handler {
	asGeneric := genericHandler(func(w http.ResponseWriter, req *http.Request) (interface{}, error) {
		return nil, next(w, req)
	})
	wrapped := applyWrappers(asGeneric, more)
	return wrapForErrorWriting(wrapped)
}

func Wrap(next func(*http.Request) (interface{}, error), more ...EndpointWrapper) Handler {
	asGeneric := genericHandler(func(w http.ResponseWriter, req *http.Request) (interface{}, error) {
		return next(req)
	})
	wrapped := applyWrappers(asGeneric, more)
	return wrapForResponseAndErrorWriting(wrapped)
}

func ReadBody(next func(req *http.Request, data []byte) (interface{}, error), more ...EndpointWrapper) Handler {
	asGeneric := genericHandler(func(w http.ResponseWriter, req *http.Request) (interface{}, error) {
		defer req.Body.Close()

		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, HttpCode(http.StatusBadRequest)
		}
		if len(data) > MaxRequestDataBytes {
			return nil, WithCode{http.StatusRequestEntityTooLarge, fmt.Sprintf("Too large (%d > %d)", len(data), MaxRequestDataBytes)}
		}
		return next(req, data)
	})
	wrapped := applyWrappers(asGeneric, more)

	return wrapForResponseAndErrorWriting(func(w http.ResponseWriter, req *http.Request) (interface{}, error) {
		return wrapped(w, req)
	})
}

func DiscardBody(next func(req *http.Request) (interface{}, error), more ...EndpointWrapper) Handler {
	asGeneric := genericHandler(func(w http.ResponseWriter, req *http.Request) (interface{}, error) {
		defer req.Body.Close()
		io.Copy(ioutil.Discard, req.Body)

		return next(req)
	})
	wrapped := applyWrappers(asGeneric, more)

	return wrapForResponseAndErrorWriting(func(w http.ResponseWriter, req *http.Request) (interface{}, error) {
		return wrapped(w, req)
	})
}
