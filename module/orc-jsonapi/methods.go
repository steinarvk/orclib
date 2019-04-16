package jsonapi

import "net/http"

type Handler func(w http.ResponseWriter, req *http.Request) error

type Methods struct {
	Get    Handler
	Post   Handler
	Put    Handler
	Delete Handler
	Patch  Handler
	Others map[string]Handler
}

func (m Methods) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	_ = m.makeHandler()(w, req)
}

func (m Methods) makeHandler() Handler {
	callHandler := func(next Handler, w http.ResponseWriter, req *http.Request) error {
		return next(w, req)
	}
	return func(w http.ResponseWriter, req *http.Request) error {
		switch req.Method {
		case "GET":
			if handler := m.Get; handler != nil {
				return callHandler(handler, w, req)
			}
		case "POST":
			if handler := m.Post; handler != nil {
				return callHandler(handler, w, req)
			}
		case "PUT":
			if handler := m.Put; handler != nil {
				return callHandler(handler, w, req)
			}
		case "DELETE":
			if handler := m.Delete; handler != nil {
				return callHandler(handler, w, req)
			}
		case "PATCH":
			if handler := m.Patch; handler != nil {
				return callHandler(handler, w, req)
			}
		}

		if m.Others != nil {
			handler, ok := m.Others[req.Method]
			if ok && handler != nil {
				return callHandler(handler, w, req)
			}
		}

		writeResponse(w, http.StatusMethodNotAllowed, nil)
		return HttpCode(http.StatusMethodNotAllowed)
	}
}
