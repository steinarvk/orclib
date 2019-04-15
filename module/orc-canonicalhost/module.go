package canonicalhost

import (
	"github.com/steinarvk/orc"

	identity "github.com/steinarvk/orclib/module/orc-identity"
	server "github.com/steinarvk/orclib/module/orc-server"
)

var CanonicalHost string

type Module struct {
	CanonicalHost string
}

func (m *Module) ModuleName() string { return "CanonicalHost" }

var M = &Module{}

func (m *Module) OnRegister(hooks orc.ModuleHooks) {
	hooks.OnUse(func(ctx orc.UseContext) {
		ctx.Use(server.M)
		ctx.Flags.StringVar(&m.CanonicalHost, "canonical_host", "", "canonical hostname of this server, for use as identity")
	})
	hooks.OnSetup(func() error {
		if m.CanonicalHost == "" {
			m.CanonicalHost = server.M.Addr
		}

		identity.MutateIdentity(func(claim *identity.IdentityClaim) {
			claim.CanonicalHost = m.CanonicalHost
		})

		CanonicalHost = m.CanonicalHost
		return nil
	})
}
