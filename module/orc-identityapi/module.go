package identityapi

import (
	"fmt"
	"net/http"

	"github.com/steinarvk/orc"

	"github.com/steinarvk/orclib/lib/cryptopacket"
	"github.com/steinarvk/orclib/module/orc-debug"

	orcephemeralkeys "github.com/steinarvk/orclib/module/orc-ephemeralkeys"
	httprouter "github.com/steinarvk/orclib/module/orc-httprouter"
	identity "github.com/steinarvk/orclib/module/orc-identity"
	jsonapi "github.com/steinarvk/orclib/module/orc-jsonapi"
	orcpersistentkeys "github.com/steinarvk/orclib/module/orc-persistentkeys"
)

type Module struct {
}

var M = &Module{}

func (m *Module) ModuleName() string { return "IdentityAPI" }

func (m *Module) OnRegister(hooks orc.ModuleHooks) {
	hooks.OnUse(func(u orc.UseContext) {
		u.Use(jsonapi.M)
		u.Use(orcdebug.M)
	})

	hooks.OnSetup(func() error {
		return nil
	})

	hooks.OnStart(func() error {
		getIdentityHandler := jsonapi.Methods{
			Get: jsonapi.DiscardBody(func(req *http.Request) (interface{}, error) {
				keyset := orcpersistentkeys.M.Keys
				if keyset == nil {
					keyset = orcephemeralkeys.M.EphemeralKeys
				}
				if keyset == nil {
					return identity.Identity(), nil
				}
				packet, err := cryptopacket.PackUnencryptedJSON(identity.Identity(), keyset)
				if err != nil {
					return nil, fmt.Errorf("Error packing keys: %v", err)
				}
				return packet, nil
			}),
		}
		jsonapi.M.Handle("/internal/identity", getIdentityHandler)
		httprouter.M.HandleDebug("/internal-api/identity", getIdentityHandler)
		return nil
	})
}
