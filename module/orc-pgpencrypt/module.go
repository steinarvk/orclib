package orcpgpencrypt

import (
	"github.com/steinarvk/orc"
	"github.com/steinarvk/orclib/lib/pgpencrypt"
)

type Module struct {
	Config    pgpencrypt.Config
	Encrypter *pgpencrypt.Encrypter
}

func (m *Module) ModuleName() string { return "PGPEncrypt" }

var M = &Module{}

func (m *Module) OnRegister(hooks orc.ModuleHooks) {
	hooks.OnUse(func(ctx orc.UseContext) {
		ctx.Flags.StringSliceVar(&m.Config.KeyringFilenames, "pgp_keyring_file", nil, "keyring(s) to read from file for PGP encryption")
		ctx.Flags.StringSliceVar(&m.Config.KeyringURLs, "pgp_keyring_url", nil, "keyring(s) to read from HTTPS URLs for PGP encryption")
		ctx.Flags.StringSliceVar(&m.Config.Recipients, "pgp_encrypt_to", nil, "recipient(s) identifiers (name or email associated with key)")
	})

	hooks.OnStart(func() error {
		outputEncrypter, err := pgpencrypt.FromConfig(&m.Config)
		if err != nil {
			return err
		}
		m.Encrypter = outputEncrypter
		return nil
	})
}
