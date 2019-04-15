package authinterface

import (
	"fmt"
	"net/http"
	"strings"
)

type AnyOfGatekeepers []Gatekeeper

func (a AnyOfGatekeepers) CheckAuth(r *http.Request) (*AuthSuccessInfo, error) {
	var lastErr error
	for _, x := range a {
		success, err := x.CheckAuth(r)
		if err == nil {
			return success, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func (a AnyOfGatekeepers) DemandAuth(w http.ResponseWriter) error {
	for _, x := range a {
		if err := x.DemandAuth(w); err != nil {
			return err
		}
	}
	return nil
}

func (a AnyOfGatekeepers) GatekeeperDescription() string {
	var rv []string
	for _, x := range a {
		rv = append(rv, x.GatekeeperDescription())
	}
	return fmt.Sprintf("AnyOf[%s]", strings.Join(rv, ","))
}
