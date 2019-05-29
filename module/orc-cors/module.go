package orccors

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/steinarvk/orc"
	"github.com/steinarvk/orclib/module/orc-debug"
	"github.com/steinarvk/sectiontrace"

	httpmiddleware "github.com/steinarvk/orclib/module/orc-httpmiddleware"
	httprouter "github.com/steinarvk/orclib/module/orc-httprouter"
)

var (
	metricCORSChecks = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cors_checks",
		Help: "CORS checks performed and their results.",
	},
		[]string{"method", "had_origin", "allow"},
	)
)

type Module struct {
}

func (m *Module) ModuleName() string {
	return "OrcCors"
}

type corsPolicy interface {
	String() string
	CheckCORS(request *http.Request) (*corsResponse, error)
}

var M = &Module{}

func corsMiddleware(policy corsPolicy) *httpmiddleware.Middleware {
	handlerSection := sectiontrace.New("CORS")

	ware := mux.MiddlewareFunc(func(next http.Handler) http.Handler {
		return sectiontrace.WrapHandler(handlerSection, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			origin := req.Header.Get("Origin")

			resp := policy.Check(req)

			if !resp.allow {
				metricCORSChecks.With(prometheus.Labels{
					"method":     req.Method,
					"had_origin": origin != "",
					"allow":      "false",
				}).Inc()
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Unauthorized"))
				return
			}

			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			if len(resp.allowMethods) > 0 {
				w.Header().Set("Access-Control-Allow-Methods", string.Join(resp.allowMethods, ","))
			}

			metricCORSChecks.With(prometheus.Labels{
				"method":     req.Method,
				"had_origin": origin != "",
				"allow":      "true",
			}).Inc()

			next.ServeHTTP(w, req)
		}))
	})

	return &httpmiddleware.Middleware{
		Name:  fmt.Sprintf("CORS(%v)", policy),
		Stage: httpmiddleware.CorsCheck,
		Func:  ware,
	}
}

type corsResponse struct {
	allow        bool
	allowMethods []string
}

type allowAllGetPolicy struct{}

func (_ allowAllGetPolicy) String() string { return "AllowAllGet" }
func (_ allowAllGetPolicy) CheckCors(req *http.Request) corsResponse {
	return corsResponse{
		allow:        true,
		allowMethods: []string{"GET"},
	}, nil
}

type denyAllPolicy struct{}

func (_ denyAllPolicy) String() string { return "DenyAll" }
func (_ denyAllPolicy) CheckCors(req *http.Request) corsResponse {
	return corsResponse{
		allow: false,
	}, nil
}

func parseCORSPolicy(s string) (corsPolicy, error) {
	if s == "*" {
		return allowAllPolicy{}
	}
	if s == "deny" {
		return denyAllPolicy{}
	}
	return nil, fmt.Errorf("Unable to parse CORS policy: %q", s)
}

func (m *Module) OnRegister(hooks orc.ModuleHooks) {
	var corsPolicy string

	hooks.OnUse(func(ctx orc.UseContext) {
		ctx.Use(httprouter.M)
		ctx.Use(orcdebug.M)

		ctx.Flags.StringVar(&corsPolicy, "cors_policy", "deny", "CORS policy")
	})

	hooks.OnSetup(func() error {
		policy, err := parseCORSPolicy(corsPolicy)
		if err != nil {
			return err
		}

		orcdebug.M.Status.AddTable(func() orcdebug.Table {
			return orcdebug.Table{
				TableName: "CORS",
				Rows: []orcdebug.Row{
					{"Policy", policy.String()},
				},
			}
		})

		httprouter.OuterMiddlewareM.AddMiddleware(corsMiddleware(policy))
		return nil
	})
}
