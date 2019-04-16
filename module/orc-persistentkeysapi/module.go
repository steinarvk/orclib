package persistentkeysapi

import (
	"fmt"
	"net/http"

	"github.com/steinarvk/orc"

	"github.com/steinarvk/orclib/lib/cryptopacket"
	"github.com/steinarvk/orclib/module/orc-debug"

	canonicalhost "github.com/steinarvk/orclib/module/orc-canonicalhost"
	httprouter "github.com/steinarvk/orclib/module/orc-httprouter"
	jsonapi "github.com/steinarvk/orclib/module/orc-jsonapi"
	orckeys "github.com/steinarvk/orclib/module/orc-persistentkeys"
)

type Module struct {
}

var M = &Module{}

func (m *Module) ModuleName() string { return "KeysAPI" }

func (m *Module) OnRegister(hooks orc.ModuleHooks) {
	hooks.OnUse(func(u orc.UseContext) {
		u.Use(jsonapi.M)
		u.Use(canonicalhost.M)
		u.Use(orckeys.M)
		u.Use(orcdebug.M)
	})

	hooks.OnSetup(func() error {
		return nil
	})

	hooks.OnStart(func() error {
		getPublicKeysHandler := jsonapi.Methods{
			Get: jsonapi.DiscardBody(func(req *http.Request) (interface{}, error) {
				if orckeys.M.Keys == nil {
					return nil, jsonapi.WithCode{404, "No keys present"}
				}
				publicKeys := orckeys.M.Keys.Public()
				packet, err := cryptopacket.PackUnencryptedJSON(publicKeys, orckeys.M.Keys)
				if err != nil {
					return nil, fmt.Errorf("Error packing keys: %v", err)
				}
				return packet, nil
			}),
		}
		jsonapi.M.Handle("/internal/public-keys", getPublicKeysHandler)
		httprouter.M.HandleDebug("/internal-api/public-keys", getPublicKeysHandler)
		return nil
	})
}
