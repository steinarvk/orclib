package httpmiddleware

import "github.com/gorilla/mux"

type Stage int

const (
	RequestReceived = Stage(1000)
	Auth            = Stage(2000)
	RequestAccepted = Stage(3000)
)

func (s Stage) Do(name string, f mux.MiddlewareFunc) *Middleware {
	return &Middleware{name, s, f}
}
