package orctinkgcpkms

import (
	"github.com/google/tink/go/core/registry"
	"github.com/google/tink/go/integration/gcpkms"
	"github.com/sirupsen/logrus"
	"github.com/steinarvk/orc"
)

type Module struct {
}

func (m *Module) ModuleName() string { return "TinkGCPKMS" }

var M = &Module{}

func (m *Module) OnRegister(hooks orc.ModuleHooks) {
	hooks.OnUse(func(ctx orc.UseContext) {
	})

	hooks.OnSetup(func() error {
		const (
			prefix = "gcp-kms://"
		)
		client := &gcpkms.GCPClient{}
		_, err := client.LoadDefaultCredentials()
		if err != nil {
			logrus.Infof("GCP KMS: %v", err)
			return nil
		}
		registry.RegisterKMSClient(client)
		return nil
	})
}
