package orcclient

import (
	"crypto/x509"
	"net/http"
	"sync"
	"time"

	"github.com/steinarvk/orc"
	"github.com/steinarvk/orclib/lib/authinterface"
	orcouterauth "github.com/steinarvk/orclib/module/orc-outerauth"
	trustedcerts "github.com/steinarvk/orclib/module/orc-trustedcerts"
)

type Config struct {
	RootCAs                     *x509.CertPool
	AllowOutboundOnlyToSuffixes []string
	OutboundRequestTimeout      time.Duration
}

type Module struct {
	cfg Config

	mu           sync.Mutex
	authProvider authinterface.AuthProvider
}

var M = &Module{}

func (m *Module) ModuleName() string { return "OrcClient" }

func (m *Module) OnRegister(hooks orc.ModuleHooks) {
	hooks.OnUse(func(u orc.UseContext) {
		u.Use(trustedcerts.M)
		u.Use(orcouterauth.M)

		u.Flags.StringSliceVar(&m.cfg.AllowOutboundOnlyToSuffixes, "orcclient_allowed_outbound_domains", nil, "if nonempty, restrict orcclient connections to hosts with one of the given suffixes")
		u.Flags.DurationVar(&m.cfg.OutboundRequestTimeout, "orcclient_request_timeout", 10*time.Second, "timeout for outbound requests")
	})

	hooks.OnStart(func() error {
		m.cfg.RootCAs = trustedcerts.M.RootCAs

		return nil
	})
}

func (m *Module) createClient() (*http.Client, error) {
	return m.cfg.createClient()
}

func (m *Module) isHostAllowed(host string) bool {
	return m.cfg.isHostAllowed(host)
}

func (m *Module) addAuth(req *http.Request, targetCanonicalHost string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if p := orcouterauth.M.Provider; p != nil {
		reqCtx := authinterface.RequestContext{
			RecipientHost: targetCanonicalHost,
			RequestTime:   time.Now(),
		}
		headers, err := p.MakeAuthHeaders(reqCtx)
		if err != nil {
			return err
		}
		for headerKey, headerValue := range headers {
			req.Header.Set(headerKey, headerValue)
		}
	}

	return nil
}
