package orcprometheus

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/steinarvk/orc"
	httprouter "github.com/steinarvk/orclib/module/orc-httprouter"
)

type Module struct{}

var M = &Module{}

func (m *Module) ModuleName() string {
	return "OrcPrometheus"
}

func (m *Module) OnRegister(hooks orc.ModuleHooks) {
	hooks.OnUse(func(ctx orc.UseContext) {
		ctx.Use(httprouter.M)
	})

	hooks.OnStart(func() error {
		httprouter.M.HandleMetrics(promhttp.Handler())
		return nil
	})
}
