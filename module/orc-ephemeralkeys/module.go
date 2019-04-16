package orcephemeralkeys

import (
	"fmt"

	"github.com/steinarvk/orc"
	"github.com/steinarvk/orclib/lib/orckeys"
	canonicalhost "github.com/steinarvk/orclib/module/orc-canonicalhost"
	identity "github.com/steinarvk/orclib/module/orc-identity"
)

type Module struct {
	Owner         string
	EphemeralKeys *orckeys.Keys
}

func (m *Module) ModuleName() string { return "EphemeralKeys" }

var M = &Module{}

func (m *Module) OnRegister(hooks orc.ModuleHooks) {
	hooks.OnUse(func(ctx orc.UseContext) {
		ctx.Use(canonicalhost.M)
	})

	hooks.OnStart(func() error {
		m.Owner = fmt.Sprintf("%s::%s", canonicalhost.M.CanonicalHost, "ephemeral")
		keys, err := orckeys.Generate(m.Owner)
		if err != nil {
			return err
		}

		m.EphemeralKeys = keys

		identity.MutateIdentity(func(claim *identity.IdentityClaim) {
			publicKeys := keys.Public()
			claim.EphemeralPublicKeys = &publicKeys
		})

		return nil
	})
}
