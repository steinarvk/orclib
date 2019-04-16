package jsonapi

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/steinarvk/orc"

	httprouter "github.com/steinarvk/orclib/module/orc-httprouter"
)

type Module struct {
	apimux *mux.Router
}

var M = &Module{}

func (m *Module) ModuleName() string { return "JSONAPI" }

func (m *Module) OnRegister(hooks orc.ModuleHooks) {
	hooks.OnUse(func(u orc.UseContext) {
		u.Use(httprouter.M)

		u.Flags.IntVar(&MaxRequestDataBytes, "max_data_bytes", MaxRequestDataBytes, "max data bytes to accept in API requests")
	})

	hooks.OnStart(func() error {
		m.apimux = httprouter.M.MainRouter.PathPrefix("/api/").Subrouter()
		m.apimux.NotFoundHandler = http.HandlerFunc(EndpointNotFoundHandler)
		return nil
	})
}
