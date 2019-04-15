package httprouter

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/gorilla/mux"
	"github.com/steinarvk/orc"
	"github.com/steinarvk/sectiontrace"

	httpmiddleware "github.com/steinarvk/orclib/module/orc-httpmiddleware"
)

var (
	OuterMiddlewareM   = httpmiddleware.NewModule("Outer")
	DebugMiddlewareM   = httpmiddleware.NewModule("Debug")
	MetricsMiddlewareM = httpmiddleware.NewModule("Debug")
	MainMiddlewareM    = httpmiddleware.NewModule("Main")

	middlewareModules = []orc.Module{
		OuterMiddlewareM,
		DebugMiddlewareM,
		MetricsMiddlewareM,
		MainMiddlewareM,
	}
)

type Module struct {
	metricsWrappedHandler http.Handler
	debugWrappedHandler   http.Handler
	mainWrappedHandler    http.Handler

	debugRouter   *mux.Router
	metricsRouter *mux.Router
	MainRouter    *mux.Router

	debugHandlers []string

	flagExposeTraceContexts bool
}

type HandlerType struct {
	Debug   bool
	Main    bool
	Metrics bool
}

var M = &Module{}

func (m *Module) MakeHandler(handlerName string, ht HandlerType) http.Handler {
	r := mux.NewRouter()
	if ht.Metrics {
		r.Path("/metrics").Handler(m.metricsWrappedHandler)
	}

	if ht.Debug {
		for _, alias := range DebugAliases {
			r.PathPrefix(alias).Handler(m.debugWrappedHandler)
		}
	}

	if ht.Main {
		r.PathPrefix("/").Handler(m.mainWrappedHandler)
	}

	wrapped := OuterMiddlewareM.Wrap(r)

	handlerSection := sectiontrace.New(fmt.Sprintf("HTTPHandler(%s)", handlerName))

	realOuterHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx, sec := handlerSection.Begin(req.Context())
		defer func() { sec.End(nil) }()

		if m.flagExposeTraceContexts {
			beginRec := sec.GetBeginRecord()
			w.Header().Set("X-Orc-Trace", fmt.Sprintf("%s/%d", beginRec.Scope, beginRec.ID))
		}

		req = req.WithContext(ctx)

		wrapped.ServeHTTP(w, req)
	})

	return realOuterHandler
}

var (
	DebugAliases = []string{"/debug/", "/_/"}
)

func (m *Module) HandleMetrics(handler http.Handler) {
	m.metricsRouter.Path("/metrics").Handler(handler)
	m.HandleDebug("metrics", handler)
}

func (m *Module) ListDebugHandlers() []string {
	return m.debugHandlers
}

func (m *Module) HandleDebug(path string, handler http.Handler) {
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	for _, alias := range DebugAliases {
		m.debugRouter.Path(alias + path).Handler(handler)
	}
	m.debugHandlers = append(m.debugHandlers, DebugAliases[0]+path)
	sort.Strings(m.debugHandlers)
}

func (m *Module) setupRouters() {
	m.debugRouter = mux.NewRouter()
	m.metricsRouter = mux.NewRouter()
	m.MainRouter = mux.NewRouter()

	m.metricsWrappedHandler = MetricsMiddlewareM.Wrap(m.metricsRouter)
	m.debugWrappedHandler = DebugMiddlewareM.Wrap(m.debugRouter)
	m.mainWrappedHandler = MainMiddlewareM.Wrap(m.MainRouter)

	m.HandleDebug("/ping", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		w.Write([]byte("pong"))
	}))
}

func (m *Module) ModuleName() string { return "HTTPRouter" }

func (m *Module) OnRegister(hooks orc.ModuleHooks) {
	hooks.OnUse(func(u orc.UseContext) {
		for _, mwm := range middlewareModules {
			u.Use(mwm)
		}

		u.Flags.BoolVar(&m.flagExposeTraceContexts, "expose_trace_contexts", true, "expose sectiontrace contexts in a HTTP header on responses")
	})
	hooks.OnStart(func() error {
		m.setupRouters()
		return nil
	})
}
