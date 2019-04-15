package orcouterauth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/steinarvk/orclib/lib/authinterface"
	"github.com/steinarvk/orclib/lib/hashedsecret"
)

type Gatekeeper struct {
	canonicalHost string
	secrets       []*Secret
	timer         func() time.Time
}

type Provider struct {
	primarySecret *Secret
}

type Auth struct {
	Gatekeeper Gatekeeper
	Provider   Provider
}

func (p Provider) MakeAuthHeaders(ctx authinterface.RequestContext) (map[string]string, error) {
	if p.primarySecret == nil || p.primarySecret.Secret == "" {
		return nil, fmt.Errorf("No secret set")
	}
	g := &hashedsecret.Generator{
		SharedSecret:        p.primarySecret.Secret,
		TargetCanonicalHost: ctx.RecipientHost,
	}
	token, err := g.Generate(ctx.RequestTime)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		headerName: fmt.Sprintf("%s %s", headerType, token),
	}, nil
}

type Secret struct {
	Name      string `json:"name,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
	Secret    string `json:"secret"`
}

func New(canonicalHost string, filenames []string) (*Auth, error) {
	secrets, err := loadSecrets(filenames)
	if err != nil {
		return nil, err
	}

	if len(secrets) == 0 {
		return nil, fmt.Errorf("No secrets loaded")
	}

	return &Auth{
		Gatekeeper: Gatekeeper{
			canonicalHost: canonicalHost,
			secrets:       secrets,
			timer:         time.Now,
		},
		Provider: Provider{
			primarySecret: secrets[0],
		},
	}, nil
}

const (
	headerName  = "X-Authorization"
	headerType  = "OrcOuterAuth"
	maxAttempts = 3
)

func (g Gatekeeper) checkToken(token string, info authinterface.AttemptInfo) (*authinterface.AuthSuccessInfo, error) {
	for _, secret := range g.secrets {
		v := &hashedsecret.Verifier{
			CanonicalHost: g.canonicalHost,
			SharedSecret:  secret.Secret,
		}
		username := secret.Name

		ok, err := v.Verify(username, g.timer(), token)
		success := ok && (err == nil)

		if success {
			info.Username = username

			return &authinterface.AuthSuccessInfo{
				Attempt: info,
			}, nil
		}
	}

	return nil, nil
}

func (g Gatekeeper) CheckAuth(req *http.Request) (*authinterface.AuthSuccessInfo, error) {
	values, _ := req.Header[headerName]

	attempt := authinterface.AttemptInfo{
		GatekeeperID: "OrcOuterAuth",
	}

	var attemptedValues []string

	for _, value := range values {
		splitH := strings.SplitN(value, " ", 2)
		if len(splitH) < 2 {
			continue
		}
		if splitH[0] != headerType {
			continue
		}
		attemptedValues = append(attemptedValues, splitH[1])
	}

	attempt.Attempted = len(attemptedValues) > 0

	if !attempt.Attempted {
		return nil, authinterface.DenyWith(authinterface.AuthFailureInfo{Attempt: attempt})
	}

	if len(attemptedValues) > maxAttempts {
		return nil, authinterface.ErrorWith(fmt.Errorf("Too many %q header values", headerName), authinterface.AuthFailureInfo{Attempt: attempt})
	}

	for _, v := range attemptedValues {
		success, err := g.checkToken(v, attempt)
		if err != nil {
			return nil, authinterface.ErrorWith(err, authinterface.AuthFailureInfo{Attempt: attempt})
		}
		if success != nil {
			return success, nil
		}
	}

	return nil, authinterface.DenyWith(authinterface.AuthFailureInfo{Attempt: attempt})
}

func (g Gatekeeper) DemandAuth(w http.ResponseWriter) error {
	return nil
}

func (g Gatekeeper) GatekeeperDescription() string {
	var rv []string
	for _, sec := range g.secrets {
		name := "(unnamed secret)"
		if sec.Name != "" {
			name = sec.Name
		}
		if sec.Timestamp != "" {
			name += "@" + sec.Timestamp
		}
		rv = append(rv, name)
	}
	return fmt.Sprintf("HashedSecret[%s]", strings.Join(rv, ","))
}
