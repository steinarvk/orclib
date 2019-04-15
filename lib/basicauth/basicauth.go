package basicauth

import (
	"fmt"
	"net/http"

	gohttpauth "github.com/abbot/go-http-auth"
	"github.com/steinarvk/orclib/lib/authinterface"
)

type gatekeeper struct {
	gatekeeperID         string
	filename             string
	realm                string
	unsafeSecretProvider gohttpauth.SecretProvider
	unsafeBasicAuth      *gohttpauth.BasicAuth
}

func (g *gatekeeper) CheckAuth(req *http.Request) (*authinterface.AuthSuccessInfo, error) {
	unwrappedUsername, _, _ := req.BasicAuth()
	attemptInfo := authinterface.AttemptInfo{
		GatekeeperID: g.gatekeeperID,
		Username:     unwrappedUsername,
		Realm:        g.realm,
	}

	username, err := wrappingPanic(func() string {
		return g.unsafeBasicAuth.CheckAuth(req)
	})
	if username == "" || err != nil {
		failureInfo := authinterface.AuthFailureInfo{
			Attempt: attemptInfo,
		}
		if err == nil {
			return nil, authinterface.DenyWith(failureInfo)
		}
		return nil, authinterface.ErrorWith(fmt.Errorf("Error checking auth from .htpasswd file %q (for realm %q): %v", g.filename, g.realm, err), failureInfo)
	}

	attemptInfo.Username = username
	return &authinterface.AuthSuccessInfo{Attempt: attemptInfo}, nil
}

func (g *gatekeeper) GatekeeperDescription() string {
	return fmt.Sprintf("Htpasswd[%s]", g.gatekeeperID)
}

func (g *gatekeeper) DemandAuth(w http.ResponseWriter) error {
	w.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=%q", g.realm))
	return nil
}

func (g *gatekeeper) provideSecret(username string) (string, error) {
	ignoredRealm := ""
	secret, err := wrappingPanic(func() string {
		return g.unsafeSecretProvider(username, ignoredRealm)
	})
	if err != nil {
		return "", fmt.Errorf("Error reading from .htpasswd file %q (for realm %q): %v", g.filename, g.realm, err)
	}
	return secret, nil
}

func NewGatekeeperFromHtpasswdFile(filename, realm string) (authinterface.Gatekeeper, error) {
	gk := &gatekeeper{
		gatekeeperID:         fmt.Sprintf("htpasswd(%q, %q)", filename, realm),
		filename:             filename,
		unsafeSecretProvider: gohttpauth.HtpasswdFileProvider(filename),
	}
	gk.unsafeBasicAuth = gohttpauth.NewBasicAuthenticator(realm, gk.unsafeSecretProvider)

	// need to call it once to actually test that the file parses
	_, err := gk.provideSecret("")
	if err != nil {
		return nil, err
	}

	return gk, nil
}
