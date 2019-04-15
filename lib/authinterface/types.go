package authinterface

import (
	"net/http"
	"time"
)

type AttemptInfo struct {
	GatekeeperID string `json:"gatekeeper_id"`
	Username     string `json:"username,omitempty"`
	Realm        string `json:"realm,omitempty"`
}

type AuthSuccessInfo struct {
	Attempt AttemptInfo `json:"attempt"`
}

type AuthFailureInfo struct {
	Attempt AttemptInfo `json:"attempt"`
}

type RequestContext struct {
	RecipientHost string
	RequestTime   time.Time
}

type Gatekeeper interface {
	CheckAuth(r *http.Request) (*AuthSuccessInfo, error)
	DemandAuth(w http.ResponseWriter) error
	GatekeeperDescription() string
}

type AuthProvider interface {
	MakeAuthHeaders(ctx RequestContext) (map[string]string, error)
}
