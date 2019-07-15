package orcdebug

import (
	"fmt"
	"net/http"
	"os"
	"runtime/pprof"

	"github.com/steinarvk/orc"

	"github.com/steinarvk/orclib/lib/versioninfo"
	canonicalhost "github.com/steinarvk/orclib/module/orc-canonicalhost"
	httprouter "github.com/steinarvk/orclib/module/orc-httprouter"
)

type Module struct {
	Status *Status
}

var M = &Module{}

func (m *Module) RedirectMainToStatus() error {
	httprouter.M.MainRouter.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "/debug/status", http.StatusSeeOther)
	})
	return nil
}

func (m *Module) ModuleName() string { return "OrcDebug" }

func (m *Module) OnRegister(hooks orc.ModuleHooks) {
	hooks.OnUse(func(u orc.UseContext) {
		u.Use(httprouter.M)
		u.Use(canonicalhost.M)
	})

	hooks.OnSetup(func() error {
		m.Status = NewStatus()
		return nil
	})

	hooks.OnStart(func() error {
		if versioninfo.ProgramName != "" {
			entries := versioninfo.MakeJSON().(map[string]interface{})
			tbl := Table{TableName: "Version info"}
			for key, valueIntf := range entries {
				row := Row{Key: key, Value: valueIntf.(string)}
				tbl.Rows = append(tbl.Rows, row)
			}
			m.Status.AddTable(func() Table { return tbl })
		}

		httprouter.M.HandleDebug("/stacktrace", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			pprof.Lookup("goroutine").WriteTo(w, 1)
		}))
		httprouter.M.HandleDebug("/args", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Write([]byte(fmt.Sprintf("%v", os.Args)))
		}))
		httprouter.M.HandleDebug("/status", m.Status)
		httprouter.M.HandleDebug("/index", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			trustedLinkListingPageTemplate.Execute(w, httprouter.M.ListDebugHandlers())
		}))
		return nil
	})
}
