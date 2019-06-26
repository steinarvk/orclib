package orcgrpcclientcommon

import (
	"crypto/tls"

	"github.com/steinarvk/orc"
	trustedcerts "github.com/steinarvk/orclib/module/orc-trustedcerts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Module struct {
	DialOptions []grpc.DialOption
}

func (m *Module) ModuleName() string { return "gRPCClientCommon" }

var M = &Module{}

func (m *Module) OnRegister(hooks orc.ModuleHooks) {
	hooks.OnUse(func(u orc.UseContext) {
		u.Use(trustedcerts.M)
	})

	hooks.OnStart(func() error {
		var dialOpts []grpc.DialOption

		// TLS is mandatory; multiplexing on the HTTP port doesn't work.

		creds := credentials.NewTLS(&tls.Config{
			RootCAs: trustedcerts.M.RootCAs,
		})
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))

		m.DialOptions = dialOpts
		return nil
	})
}
