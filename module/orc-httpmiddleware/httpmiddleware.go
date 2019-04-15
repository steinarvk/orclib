package httpmiddleware

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/steinarvk/orc"
)

type Middleware struct {
	Name  string
	Stage Stage
	Func  mux.MiddlewareFunc
}

type Module struct {
	name       string
	middleware map[Stage][]*Middleware

	hasFlattened        bool
	flattenedMiddleware []mux.MiddlewareFunc
}

func NewModule(name string) *Module {
	return &Module{
		name: name,
	}
}

func (m *Module) Wrap(h http.Handler) http.Handler {
	if !m.hasFlattened {
		panic(fmt.Errorf("%q: Wrap() called too early: middleware not yet finalized", m.name))
	}
	for _, wrapper := range m.flattenedMiddleware {
		h = wrapper(h)
	}
	return h
}

func (m *Module) AddMiddleware(wares ...*Middleware) {
	if m.middleware == nil {
		panic(fmt.Errorf("%q: AddMiddleware called too early: not setup yet", m.name))
	}
	if m.hasFlattened {
		panic(fmt.Errorf("%q: AddMiddleware called too late: already finalized", m.name))
	}

	for _, ware := range wares {
		if ware == nil {
			continue
		}
		if ware.Stage == 0 {
			panic(fmt.Errorf("Middleware must have Stage set"))
		}
		m.middleware[ware.Stage] = append(m.middleware[ware.Stage], ware)
	}
}

func (m *Module) ModuleName() string { return fmt.Sprintf("HTTPMiddleware(%q)", m.name) }

func (m *Module) OnRegister(hooks orc.ModuleHooks) {
	hooks.OnSetup(func() error {
		m.middleware = map[Stage][]*Middleware{}
		return nil
	})

	hooks.OnStart(func() error {
		var stages []int
		for k := range m.middleware {
			stages = append(stages, int(k))
		}
		sort.Ints(stages)

		n := 1
		var flattened []mux.MiddlewareFunc

		for _, stage := range stages {
			for _, ware := range m.middleware[Stage(stage)] {
				logrus.Infof("HTTPHandler wrapper: %d [stage %d]: %q", n, stage, ware.Name)
				n++
				flattened = append(flattened, ware.Func)
			}
		}

		m.hasFlattened = true
		m.flattenedMiddleware = flattened

		return nil
	})
}
