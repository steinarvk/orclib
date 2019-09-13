package orcgrpcserver

import (
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/steinarvk/orc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	orcdebug "github.com/steinarvk/orclib/module/orc-debug"
	httprouter "github.com/steinarvk/orclib/module/orc-httprouter"
)

type Module struct {
	Server *grpc.Server
}

func (m *Module) ModuleName() string { return "gRPCServer" }

var (
	labelsGrpc    = prometheus.Labels{"grpc": "true"}
	labelsNotGrpc = prometheus.Labels{"grpc": "false"}
)

func (m *Module) hijacker(w http.ResponseWriter, req *http.Request) bool {
	grpcRequest := req.ProtoMajor == 2 && strings.HasPrefix(req.Header.Get("Content-Type"), "application/grpc")

	if !grpcRequest {
		metricRequestsInspected.With(labelsNotGrpc).Inc()
		return false
	}

	metricRequestsInspected.With(labelsGrpc).Inc()

	m.Server.ServeHTTP(w, req)

	return true
}

var M = &Module{}

func (m *Module) OnRegister(hooks orc.ModuleHooks) {
	hooks.OnUse(func(u orc.UseContext) {
		u.Use(httprouter.M)
		u.Use(orcdebug.M)
	})

	hooks.OnSetup(func() error {
		m.Server = grpc.NewServer()

		reflection.Register(m.Server)

		orcdebug.M.Status.AddTable(func() orcdebug.Table {
			return orcdebug.Table{
				TableName: "gRPC server",
				Rows: []orcdebug.Row{
					{"Active", "true"},
				},
			}
		})
		logrus.Infof("Initialized gRPC server.")
		return nil
	})

	hooks.OnStart(func() error {
		httprouter.ConnectionHijackers = append(httprouter.ConnectionHijackers, m.hijacker)
		return nil
	})
}
