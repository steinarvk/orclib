package orccors

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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
	CheckCORS(request *http.Request) corsResponse
}

var M = &Module{}

func corsMiddleware(policy corsPolicy) *httpmiddleware.Middleware {
	handlerSection := sectiontrace.New("CORS")

	ware := mux.MiddlewareFunc(func(next http.Handler) http.Handler {
		return sectiontrace.WrapHandler(handlerSection, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			origin := req.Header.Get("Origin")

			if origin == "" {
				metricCORSChecks.With(prometheus.Labels{
					"method":     req.Method,
					"had_origin": "false",
					"allow":      "true",
				}).Inc()
			} else {
				resp := policy.CheckCORS(req)

				if !resp.allow {
					metricCORSChecks.With(prometheus.Labels{
						"method":     req.Method,
						"had_origin": "true",
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
					w.Header().Set("Access-Control-Allow-Methods", strings.Join(resp.allowMethods, ","))
				}

				metricCORSChecks.With(prometheus.Labels{
					"method":     req.Method,
					"had_origin": "true",
					"allow":      "true",
				}).Inc()
			}

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
func (_ allowAllGetPolicy) CheckCORS(req *http.Request) corsResponse {
	return corsResponse{
		allow:        true,
		allowMethods: []string{"GET"},
	}
}

type denyAllPolicy struct{}

func (_ denyAllPolicy) String() string { return "DenyAll" }
func (_ denyAllPolicy) CheckCORS(req *http.Request) corsResponse {
	return corsResponse{
		allow: false,
	}
}

func parseCORSPolicy(s string) (corsPolicy, error) {
	if s == "*" {
		return allowAllGetPolicy{}, nil
	}
	if s == "deny" {
		return denyAllPolicy{}, nil
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
