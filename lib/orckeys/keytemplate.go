package orckeys

import (
	"github.com/google/tink/go/hybrid"
	"github.com/google/tink/go/signature"
)

var DefaultSigningKeyTemplate = signature.ECDSAP256KeyTemplate()
var DefaultEncryptionKeyTemplate = hybrid.ECIESHKDFAES128CTRHMACSHA256KeyTemplate()
