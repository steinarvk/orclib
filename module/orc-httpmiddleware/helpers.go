package httpmiddleware

import (
	"net/http"

	"github.com/gorilla/mux"
)

func SetHeader(key, value string) mux.MiddlewareFunc {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set(key, value)
		})
	}
}
