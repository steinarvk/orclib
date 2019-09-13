package orcouterauth

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
	"github.com/steinarvk/orc"
	"github.com/steinarvk/orclib/lib/authinterface"
	orcbasicauth "github.com/steinarvk/orclib/lib/basicauth"
	"github.com/steinarvk/orclib/lib/orcouterauth"
	"github.com/steinarvk/orclib/module/orc-debug"
	"github.com/steinarvk/sectiontrace"

	canonicalhost "github.com/steinarvk/orclib/module/orc-canonicalhost"
	httpmiddleware "github.com/steinarvk/orclib/module/orc-httpmiddleware"
	httprouter "github.com/steinarvk/orclib/module/orc-httprouter"
)

var (
	DefaultDisableInboundAuth bool = false
)

var (
	metricOuterAuthRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "outer_auth_requests",
		Help: "Number of requests for which outer auth was processed.",
	},
		[]string{"realm", "status"},
	)
)

type Module struct {
	Provider authinterface.AuthProvider
}

func (m *Module) ModuleName() string {
	return "OrcOuterAuth"
}

var M = &Module{}

func gatekeeperMiddleware(rootName string, gk authinterface.Gatekeeper) *httpmiddleware.Middleware {
	handlerSection := sectiontrace.New("Gatekeeper")

	ware := mux.MiddlewareFunc(func(next http.Handler) http.Handler {
		return sectiontrace.WrapHandler(handlerSection, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			labels := prometheus.Labels{
				"realm":  rootName,
				"status": "null",
			}
			defer func() {
				metricOuterAuthRequests.With(labels).Inc()
			}()

			success, err := gk.CheckAuth(req)
			if err != nil {
				labels["status"] = "failed"

				if failure, ok := err.(authinterface.ErrorWithFailureInfo); ok {
					attempt := failure.AuthFailureInfo().Attempt
					logrus.WithFields(logrus.Fields{
						"root":          rootName,
						"gatekeeper_id": attempt.GatekeeperID,
						"username":      attempt.Username,
						"realm":         attempt.Realm,
						"error":         err,
					}).Infof("Auth failed")
				} else {
					logrus.WithFields(logrus.Fields{
						"root":  rootName,
						"error": err,
					}).Infof("Auth failed with no AuthFailureInfo")
				}
			} else {
				logrus.WithFields(logrus.Fields{
					"root":          rootName,
					"gatekeeper_id": success.Attempt.GatekeeperID,
					"username":      success.Attempt.Username,
					"realm":         success.Attempt.Realm,
				}).Infof("Auth successful")
			}

			if success == nil || err != nil {
				if err := gk.DemandAuth(w); err != nil {
					logrus.WithFields(logrus.Fields{
						"root":  rootName,
						"error": err,
					}).Error("Failed to demand auth")
				}
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Unauthorized"))
				return
			}
			labels["status"] = "ok"
			next.ServeHTTP(w, req)
		}))
	})

	return &httpmiddleware.Middleware{
		Name:  fmt.Sprintf("Gatekeeper(%q, %q)", rootName, gk.GatekeeperDescription()),
		Stage: httpmiddleware.Auth,
		Func:  ware,
	}
}

func loadHtpasswdGatekeeper(realm string, htpasswds []string) (authinterface.Gatekeeper, error) {
	if len(htpasswds) == 0 {
		logrus.WithFields(logrus.Fields{
			"realm": realm,
		}).Warnf("No htpasswd provided: denying all")
		return authinterface.DenyAll, nil
	}

	var gks []authinterface.Gatekeeper

	for _, htpasswd := range htpasswds {
		gk, err := orcbasicauth.NewGatekeeperFromHtpasswdFile(htpasswd, realm)
		if err != nil {
			return nil, err
		}
		gks = append(gks, gk)
	}

	return authinterface.AnyOfGatekeepers(gks), nil
}

func (m *Module) OnRegister(hooks orc.ModuleHooks) {
	var debugHtpasswds []string
	var metricsHtpasswds []string
	var outerAuthConfigs []string
	var disableInboundOuterAuth bool
	var disableInboundDebugOuterAuth bool

	hooks.OnUse(func(ctx orc.UseContext) {
		ctx.Use(httprouter.M)
		ctx.Use(canonicalhost.M)
		ctx.Use(orcdebug.M)

		ctx.Flags.StringSliceVar(&outerAuthConfigs, "outer_auth", nil, "outer auth configuration for inbound and outbound requests")
		ctx.Flags.BoolVar(&disableInboundOuterAuth, "disable_inbound_outer_auth", DefaultDisableInboundAuth, "disable outer auth for inbound requests (allowing all instead) for main")
		ctx.Flags.BoolVar(&disableInboundDebugOuterAuth, "disable_inbound_debug_outer_auth", DefaultDisableInboundAuth, "disable outer auth for inbound requests (allowing all instead) for debug")

		ctx.Flags.StringSliceVar(&debugHtpasswds, "debug_htpasswd", nil, "htpasswd file for inbound debug requests")
		ctx.Flags.StringSliceVar(&metricsHtpasswds, "metrics_htpasswd", nil, "htpasswd file for inbound metrics requests")
	})

	hooks.OnSetup(func() error {
		debugRealm := fmt.Sprintf("%s (%s)", canonicalhost.CanonicalHost, "debug")
		metricsRealm := fmt.Sprintf("%s (%s)", canonicalhost.CanonicalHost, "metrics")

		debugAuth, err := loadHtpasswdGatekeeper(debugRealm, debugHtpasswds)
		if err != nil {
			return err
		}

		metricsAuth, err := loadHtpasswdGatekeeper(metricsRealm, metricsHtpasswds)
		if err != nil {
			return err
		}

		var outerAuthProvider authinterface.AuthProvider
		mainOuterAuth := authinterface.DenyAll

		if len(outerAuthConfigs) > 0 {
			auth, err := orcouterauth.New(canonicalhost.CanonicalHost, outerAuthConfigs)
			if err != nil {
				return err
			}
			mainOuterAuth = auth.Gatekeeper
			outerAuthProvider = auth.Provider
			m.Provider = outerAuthProvider
		}

		if disableInboundOuterAuth {
			mainOuterAuth = authinterface.AllowAll
		}

		if disableInboundDebugOuterAuth {
			debugAuth = authinterface.AllowAll
		}

		outgoingDesc := "(none)"
		if outerAuthProvider != nil {
			outgoingDesc = "(set)"
		}

		orcdebug.M.Status.AddTable(func() orcdebug.Table {
			return orcdebug.Table{
				TableName: "Auth",
				Rows: []orcdebug.Row{
					{"Main (outer)", mainOuterAuth.GatekeeperDescription()},
					{"Debug", debugAuth.GatekeeperDescription()},
					{"Metrics", metricsAuth.GatekeeperDescription()},
					{"(outer auth outgoing)", outgoingDesc},
				},
			}
		})

		httprouter.MainMiddlewareM.AddMiddleware(gatekeeperMiddleware("main", mainOuterAuth))
		httprouter.DebugMiddlewareM.AddMiddleware(gatekeeperMiddleware("debug", debugAuth))
		httprouter.MetricsMiddlewareM.AddMiddleware(gatekeeperMiddleware("metrics", metricsAuth))
		return nil
	})
}
