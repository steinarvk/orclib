package publickeyregistry

import (
	"fmt"

	"github.com/steinarvk/orclib/lib/orckeys"
)

func (m *Module) LookupPublicKeys(canonicalName string) (*orckeys.PublicKeyPacket, error) {
	if m.PublicKeys != nil {
		value, ok := m.PublicKeys[canonicalName]
		if ok {
			return value, nil
		}
	}
	return nil, fmt.Errorf("No public keys found for %q", canonicalName)
}
