package publickeyregistry

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/steinarvk/orc"
	"github.com/steinarvk/orclib/lib/orckeys"
)

type Module struct {
	PublicKeys map[string]*orckeys.PublicKeyPacket
}

func (m *Module) ModuleName() string { return "ServerKeys" }

var M = &Module{}

func (m *Module) OnRegister(hooks orc.ModuleHooks) {
	var publicKeysFilename string

	hooks.OnUse(func(ctx orc.UseContext) {
		ctx.Flags.StringVar(&publicKeysFilename, "public_keys_filename", "", "name of JSON file from which to load server keys")
	})

	hooks.OnSetup(func() error {
		m.PublicKeys = map[string]*orckeys.PublicKeyPacket{}
		return nil
	})

	hooks.OnStart(func() error {
		if publicKeysFilename == "" {
			return nil
		}

		data, err := ioutil.ReadFile(publicKeysFilename)
		if err != nil {
			return fmt.Errorf("unable to open --public_keys_filename=%q: %v", publicKeysFilename, err)
		}

		var publicKeysFromFile map[string]*orckeys.PublicKeyPacket

		if err := json.Unmarshal(data, &publicKeysFromFile); err != nil {
			return fmt.Errorf("unable to parse --public_keys_filename=%q: %v", publicKeysFilename, err)
		}

		if publicKeysFromFile != nil {
			for k := range publicKeysFromFile {
				m.PublicKeys[k] = publicKeysFromFile[k]
			}
		}

		return nil
	})
}
