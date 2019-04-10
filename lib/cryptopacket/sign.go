package cryptopacket

import (
	"encoding/base64"
	"fmt"

	"github.com/google/tink/go/tink"
	"github.com/steinarvk/orclib/lib/canonicalgojson"
)

func SignString(sgn tink.Signer, data string) (string, error) {
	return signString(sgn, data)
}

func SignJSON(sgn tink.Signer, data interface{}) (string, error) {
	canonical, err := canonicalgojson.MarshalCanonicalGoJSON(data)
	if err != nil {
		return "", fmt.Errorf("Error marshalling data canonically: %v", err)
	}
	sig, err := SignString(sgn, string(canonical))
	return sig, err
}

func signString(signer tink.Signer, s string) (string, error) {
	rv, err := signer.Sign([]byte(s))
	if err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(rv), nil
}
