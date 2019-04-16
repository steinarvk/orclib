package authinterface

import "net/http"

type allowAll struct{}

var AllowAll Gatekeeper = allowAll{}

func (d allowAll) CheckAuth(r *http.Request) (*AuthSuccessInfo, error) {
	return &AuthSuccessInfo{
		Attempt: AttemptInfo{
			GatekeeperID: "AllowAll",
		},
	}, nil
}

func (d allowAll) DemandAuth(w http.ResponseWriter) error {
	return nil
}

func (d allowAll) GatekeeperDescription() string {
	return "AllowAll"
}
