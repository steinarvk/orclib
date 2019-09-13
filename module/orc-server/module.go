package server

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/steinarvk/orc"

	httprouter "github.com/steinarvk/orclib/module/orc-httprouter"
)

var Server *http.Server
var ListenAndServe func() error

type ListenAddress struct {
	Host            string
	Port            int
	NoTLSPort       int
	RandomPort      bool
	RandomNoTLSPort bool
}

type ListenSpy interface {
	ReportListening(name string, addr string)
}

type ListenPromise interface {
	GetListenAddresses() ListenAddress
}

type TLSPromise interface {
	GetTLSConfig(hostname string) (*tls.Config, error)
}

var ExternalTLS TLSPromise
var ExternalListen ListenPromise
var ExternalListenSpy ListenSpy

type Module struct {
	Addr string
}

var M = &Module{}

func (m *Module) ModuleName() string { return "Server" }

func hostnameOnly(addr string) string {
	if strings.Contains(addr, ":") {
		return strings.Split(addr, ":")[0]
	}
	return addr
}

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
	var flagNoTLSListenAddr string

	hooks.OnUse(func(ctx orc.UseContext) {
		ctx.Use(httprouter.M)

		if ExternalListen == nil {
			ctx.Flags.StringVar(&flagMetricsListenHost, "metrics_host", "", "host on which to listen for metrics, if not same as debug port")
			ctx.Flags.IntVar(&flagMetricsPort, "metrics_port", 0, "port on which to listen for debug/metrics")

			ctx.Flags.StringVar(&flagDebugListenHost, "debug_host", "", "host on which to listen for debug/metrics, if not same as main port")
			ctx.Flags.IntVar(&flagDebugPort, "debug_port", 0, "port on which to listen for debug/metrics")

			ctx.Flags.StringVar(&flagNoTLSListenAddr, "notls_addr", "", "host on which to listen without TLS")

			ctx.Flags.StringVar(&flagListenHost, "host", "", "host on which to listen")
			ctx.Flags.IntVar(&flagPort, "port", 4155, "port on which to listen")
		}

		if ExternalTLS == nil {
			ctx.Flags.StringVar(&flagServeTLSCertificate, "tls_cert", "", "TLS certificate file to use for serving")
			ctx.Flags.StringVar(&flagServeTLSKey, "tls_key", "", "TLS key file to use for serving")
		}

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

		if ExternalListen != nil {
			addrs := ExternalListen.GetListenAddresses()
			flagListenHost = addrs.Host
			flagPort = addrs.Port
			if addrs.NoTLSPort != 0 || addrs.RandomNoTLSPort {
				if addrs.RandomNoTLSPort && addrs.NoTLSPort != 0 {
					return fmt.Errorf("RandomNoTLSPort but NoTLSPort = %d", addrs.NoTLSPort)
				}
				flagNoTLSListenAddr = fmt.Sprintf("%s:%d", flagListenHost, addrs.NoTLSPort)
			}

			if addrs.Port == 0 && !addrs.RandomPort {
				return fmt.Errorf("!RandomPort but Port = 0")
			}

			m.Addr = fmt.Sprintf("%s:%d", flagListenHost, flagPort)
		}

		listenAndCallback := func(name string, addr string, serve func(lis net.Listener) error) error {
			listener, err := net.Listen("tcp", addr)
			if err != nil {
				return err
			}

			defer listener.Close()

			if ExternalListenSpy != nil {
				ExternalListenSpy.ReportListening(name, listener.Addr().String())
			}

			return serve(listener)
		}

		listenAndServeOn := func(name string, srv *http.Server, withoutTLS bool) error {
			switch {
			case withoutTLS:
				logrus.Infof("Serving raw HTTP on %q", srv.Addr)
				return listenAndCallback(name, srv.Addr, func(lis net.Listener) error {
					return srv.Serve(lis)
				})
			case ExternalTLS != nil:
				tlsConfig, err := ExternalTLS.GetTLSConfig(hostnameOnly(srv.Addr))
				if err != nil {
					return err
				}

				tlsConfig.NextProtos = append(tlsConfig.NextProtos, "h2", "http/1.1")

				if tlsConfig == nil {
					logrus.Infof("Serving raw HTTP on %q", srv.Addr)
					return listenAndCallback(name, srv.Addr, func(lis net.Listener) error {
						return srv.Serve(lis)
					})
				}

				logrus.Infof("Serving TLS on %q (configuration not from file)", srv.Addr)

				listener, err := tls.Listen("tcp", srv.Addr, tlsConfig)
				if err != nil {
					return err
				}
				defer listener.Close()

				if ExternalListenSpy != nil {
					ExternalListenSpy.ReportListening(name, listener.Addr().String())
				}

				return srv.Serve(listener)

			case flagServeTLSCertificate != "" || flagServeTLSKey != "":
				logrus.Infof("Serving TLS on %q from certificate=%q key=%q", srv.Addr, flagServeTLSCertificate, flagServeTLSKey)
				return listenAndCallback(name, srv.Addr, func(lis net.Listener) error {
					return srv.ServeTLS(lis, flagServeTLSCertificate, flagServeTLSKey)
				})
			default:
				logrus.Infof("Serving raw HTTP on %q", srv.Addr)
				return listenAndCallback(name, srv.Addr, func(lis net.Listener) error {
					return srv.Serve(lis)
				})
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
				if err := listenAndServeOn("debug", debugServer, false); err != nil {
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
				if err := listenAndServeOn("metrics", metricsServer, false); err != nil {
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
			if flagNoTLSListenAddr != "" {
				serverCopy := *mainServer
				serverCopy.Addr = flagNoTLSListenAddr
				go func() {
					if err := listenAndServeOn("notls", &serverCopy, true); err != nil {
						logrus.Infof("NoTLS server shut down with error: %v", err)
					}
				}()
			}

			return listenAndServeOn("main", mainServer, false)
		}

		return nil
	})
}
