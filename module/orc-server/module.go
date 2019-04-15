package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/steinarvk/orc"

	httprouter "github.com/steinarvk/orclib/module/orc-httprouter"
)

var Server *http.Server
var ListenAndServe func() error

type Module struct {
	Addr string
}

var M = &Module{}

func (m *Module) ModuleName() string { return "Server" }

func (m *Module) OnRegister(hooks orc.ModuleHooks) {
	var flagListenHost string
	var flagPort int
	var flagReadTimeout time.Duration
	var flagWriteTimeout time.Duration
	var flagMaxHeaderBytes int
	var flagServeTLSCertificate string
	var flagServeTLSKey string
	var flagDebugPort int
	var flagDebugListenHost string
	var flagMetricsPort int
	var flagMetricsListenHost string

	hooks.OnUse(func(ctx orc.UseContext) {
		ctx.Use(httprouter.M)

		ctx.Flags.StringVar(&flagMetricsListenHost, "metrics_host", "", "host on which to listen for metrics, if not same as debug port")
		ctx.Flags.IntVar(&flagMetricsPort, "metrics_port", 0, "port on which to listen for debug/metrics")

		ctx.Flags.StringVar(&flagDebugListenHost, "debug_host", "", "host on which to listen for debug/metrics, if not same as main port")
		ctx.Flags.IntVar(&flagDebugPort, "debug_port", 0, "port on which to listen for debug/metrics")

		ctx.Flags.StringVar(&flagListenHost, "host", "", "host on which to listen")
		ctx.Flags.IntVar(&flagPort, "port", 4155, "port on which to listen")
		ctx.Flags.StringVar(&flagServeTLSCertificate, "tls_cert", "", "TLS certificate file to use for serving")
		ctx.Flags.StringVar(&flagServeTLSKey, "tls_key", "", "TLS key file to use for serving")
		ctx.Flags.IntVar(&flagMaxHeaderBytes, "max_header_bytes", 1<<20, "max header bytes to accept")
		ctx.Flags.DurationVar(&flagReadTimeout, "read_timeout", 10*time.Second, "request read timeout")
		ctx.Flags.DurationVar(&flagWriteTimeout, "write_timeout", 10*time.Second, "request write timeout")
	})

	hooks.OnSetup(func() error {
		m.Addr = fmt.Sprintf("%s:%d", flagListenHost, flagPort)
		return nil
	})

	hooks.OnStart(func() error {
		logrus.Infof("Running server on %q.", m.Addr)

		listenAndServeOn := func(srv *http.Server) error {
			if flagServeTLSCertificate != "" || flagServeTLSKey != "" {
				logrus.Infof("Serving TLS on %q from certificate=%q key=%q", srv.Addr, flagServeTLSCertificate, flagServeTLSKey)
				return srv.ListenAndServeTLS(flagServeTLSCertificate, flagServeTLSKey)
			} else {
				logrus.Infof("Serving raw HTTP on %q", srv.Addr)
				return srv.ListenAndServe()
			}
		}

		mainServer := &http.Server{
			Addr:           m.Addr,
			ReadTimeout:    flagReadTimeout,
			WriteTimeout:   flagWriteTimeout,
			MaxHeaderBytes: flagMaxHeaderBytes,
		}

		if flagDebugPort != 0 {
			debugHost := flagDebugListenHost
			if debugHost == "" {
				debugHost = flagListenHost
			}
			debugAddr := fmt.Sprintf("%s:%d", debugHost, flagDebugPort)
			debugServerValue := *mainServer
			debugServerValue.Addr = debugAddr
			debugServerValue.Handler = httprouter.M.MakeHandler("debug", httprouter.HandlerType{Debug: true})
			debugServer := &debugServerValue
			logrus.Infof("Running debug/metrics server on %q.", debugAddr)

			go func() {
				if err := listenAndServeOn(debugServer); err != nil {
					logrus.Infof("Debug/metrics server shut down with error: %v", err)
				}
			}()
		}

		if flagMetricsPort != 0 {
			metricsHost := flagMetricsListenHost
			if metricsHost == "" {
				metricsHost = flagDebugListenHost
				if metricsHost == "" {
					metricsHost = flagListenHost
				}
			}
			metricsAddr := fmt.Sprintf("%s:%d", metricsHost, flagMetricsPort)
			metricsServerValue := *mainServer
			metricsServerValue.Addr = metricsAddr
			metricsServerValue.Handler = httprouter.M.MakeHandler("metrics", httprouter.HandlerType{Metrics: true})
			metricsServer := &metricsServerValue
			logrus.Infof("Running metrics server on %q.", metricsAddr)

			go func() {
				if err := listenAndServeOn(metricsServer); err != nil {
					logrus.Infof("Metrics server shut down with error: %v", err)
				}
			}()
		}

		mainServer.Handler = httprouter.M.MakeHandler("main", httprouter.HandlerType{
			Metrics: true,
			Debug:   true,
			Main:    true,
		})

		Server = mainServer

		ListenAndServe = func() error {
			return listenAndServeOn(mainServer)
		}

		return nil
	})
}
