package authinterface

import (
	"net/http"
)

type denyAll struct{}

var DenyAll Gatekeeper = denyAll{}

func (d denyAll) CheckAuth(r *http.Request) (*AuthSuccessInfo, error) {
	return nil, DenyWith(AuthFailureInfo{
		Attempt: AttemptInfo{
			GatekeeperID: "DenyAll",
		},
	})
}

func (d denyAll) DemandAuth(w http.ResponseWriter) error {
	return nil
}

func (d denyAll) GatekeeperDescription() string {
	return "DenyAll"
}
