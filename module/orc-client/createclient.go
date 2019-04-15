package orcclient

import (
	"crypto/tls"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	canonicalhost "github.com/steinarvk/orclib/module/orc-canonicalhost"
	"golang.org/x/net/publicsuffix"
)

func (c *Config) createClient() (*http.Client, error) {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: c.RootCAs,
			},
		},
		Timeout: c.OutboundRequestTimeout,
	}, nil
}

func (c *Config) isHostAllowed(host string) bool {
	// If there's a restriction to suffixes in place, apply it.
	if len(c.AllowOutboundOnlyToSuffixes) > 0 {
		for _, suffix := range c.AllowOutboundOnlyToSuffixes {
			if strings.HasSuffix(host, suffix) {
				return true
			}
		}
		return false
	}

	// Otherwise, if we have a canonical host, require the same TLD.
	if myHost := canonicalhost.CanonicalHost; myHost != "" {
		myPublicSuffix, _ := publicsuffix.PublicSuffix(myHost)
		wantPublicSuffix, _ := publicsuffix.PublicSuffix(host)

		allow := myPublicSuffix == wantPublicSuffix
		logrus.WithFields(logrus.Fields{
			"requiredSuffix": myPublicSuffix,
			"requestedHost":  host,
			"allow":          allow,
		}).Warningf("Deciding to allow or deny outbound request based on own canonical host")
		return allow
	}

	logrus.WithFields(logrus.Fields{
		"requestedHost": host,
	}).Warningf("Allowing outbound request by default")
	return true
}
