package sectiontraceapi

import (
	"fmt"
	"net/http"

	"github.com/steinarvk/orc"
	"github.com/steinarvk/sectiontrace"

	"github.com/steinarvk/orclib/module/orc-debug"
	httprouter "github.com/steinarvk/orclib/module/orc-httprouter"
	jsonapi "github.com/steinarvk/orclib/module/orc-jsonapi"
	"github.com/steinarvk/orclib/module/orc-sectiontrace"
)

type Module struct {
}

var M = &Module{}

func (m *Module) ModuleName() string { return "SectionTracingAPI" }

func (m *Module) OnRegister(hooks orc.ModuleHooks) {
	var flagTracesToKeep int

	hooks.OnUse(func(u orc.UseContext) {
		u.Use(orcsectiontrace.M)
		u.Use(orcdebug.M)
		u.Use(jsonapi.M)

		u.Flags.IntVar(&flagTracesToKeep, "sectiontrace_buffer_size", 1000, "number of sectiontrace records retained and exposed for debugging")
	})

	hooks.OnSetup(func() error {
		return nil
	})

	hooks.OnValidate(func() error {
		if flagTracesToKeep < 0 {
			return fmt.Errorf("--sectiontrace_buffer_size: negative value invalid: %d", flagTracesToKeep)
		}
		return nil
	})

	hooks.OnStart(func() error {
		enabled := flagTracesToKeep > 0
		if enabled {
			collector := NewCollector(flagTracesToKeep)
			orcsectiontrace.M.AddCollector(collector.Collect)
			getTracesHandler := jsonapi.Methods{
				Get: jsonapi.DiscardBody(func(req *http.Request) (interface{}, error) {
					return sectiontrace.Export(collector.GetRecords()), nil
				}),
			}
			httprouter.M.HandleDebug("/traces", getTracesHandler)
		}
		return nil
	})
}
