package trustedcerts

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"github.com/steinarvk/orc"
)

type Module struct {
	RootCAs *x509.CertPool
}

func (m *Module) ModuleName() string { return "TrustedCerts" }

var M = &Module{}

func (m *Module) OnRegister(h orc.ModuleHooks) {
	var certFilenames []string

	h.OnUse(func(ctx orc.UseContext) {
		ctx.Flags.StringSliceVar(&certFilenames, "trust_tls_certs", nil, "TLS certificate files to trust for outbound connections")
	})

	h.OnStart(func() error {
		rootCAs, _ := x509.SystemCertPool()
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}

		for _, additionalFile := range certFilenames {
			certs, err := ioutil.ReadFile(additionalFile)
			if err != nil {
				return err
			}
			if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
				return fmt.Errorf("Failed to append any certificates from %q", additionalFile)
			}
		}

		m.RootCAs = rootCAs

		return nil
	})
}
