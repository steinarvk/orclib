package orcpersistentkeys

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/steinarvk/orc"
	"github.com/steinarvk/orclib/lib/orckeys"
	canonicalhost "github.com/steinarvk/orclib/module/orc-canonicalhost"
	identity "github.com/steinarvk/orclib/module/orc-identity"
	orctinkgcpkms "github.com/steinarvk/orclib/module/orc-tinkgcpkms"
)

type Module struct {
	Keys *orckeys.Keys
}

func (m *Module) ModuleName() string { return "PersistentKeys" }

var M = &Module{}

func (m *Module) OnRegister(hooks orc.ModuleHooks) {
	var keysFilename string
	var debugOrcKeysGenerate bool

	hooks.OnUse(func(ctx orc.UseContext) {
		ctx.Use(orctinkgcpkms.M)
		ctx.Use(canonicalhost.M)
		ctx.Flags.StringVar(&keysFilename, "keys_filename", "", "name of file from which to load server keys")
		ctx.Flags.BoolVar(&debugOrcKeysGenerate, "debug_generate_keys", false, "(for debugging convenience) generate new server keys that will not be persisted")
	})

	hooks.OnStart(func() error {
		if keysFilename != "" && debugOrcKeysGenerate {
			return fmt.Errorf("Cannot specify both --debug_generate_keys and --keys_filename")
		}

		if debugOrcKeysGenerate {
			keys, err := orckeys.Generate(canonicalhost.M.CanonicalHost)
			if err != nil {
				return fmt.Errorf("Error generating keys (--debug_generate_keys): %v", err)
			}

			logrus.Warningf("--debug_generate_keys: generated new keys, server has no persistent keys")

			m.Keys = keys
		} else {
			if keysFilename == "" {
				return fmt.Errorf("Missing --keys_filename (or --debug_generate_keys)")
			}

			f, err := os.Open(keysFilename)
			if err != nil {
				return fmt.Errorf("unable to open --keys_filename=%q: %v", keysFilename, err)
			}
			defer f.Close()

			keys, err := orckeys.LoadEncrypted(f, "")
			if err != nil {
				return fmt.Errorf("failed to open --keys_filename=%q: %v", keysFilename, err)
			}

			if keys.Metadata.Owner != canonicalhost.M.CanonicalHost {
				return fmt.Errorf("Loaded keys owned by %q, but canonical host is %q", keys.Metadata.Owner, canonicalhost.M.CanonicalHost)
			}

			m.Keys = keys
		}

		if m.Keys == nil {
			return fmt.Errorf("Internal error: no server keys were loaded")
		}

		identity.MutateIdentity(func(claim *identity.IdentityClaim) {
			publicKeys := m.Keys.Public()
			claim.PublicKeys = &publicKeys
		})

		return nil
	})
}
