package cryptopacket

import (
	"encoding/base64"
	"fmt"

	"github.com/google/tink/go/signature"
	"github.com/steinarvk/orclib/lib/canonicalgojson"
)

func VerifyString(data, sig, publicKey string) (bool, error) {
	dataBytes := []byte(data)

	signatureData, err := base64.RawStdEncoding.DecodeString(sig)
	if err != nil {
		return false, err
	}

	handle, err := stringToPublicKey(publicKey)
	if err != nil {
		return false, err
	}

	verifier, err := signature.NewVerifier(handle)
	if err != nil {
		return false, err
	}

	if err := verifier.Verify(signatureData, dataBytes); err != nil {
		return false, err
	}

	return true, nil
}

func VerifyJSON(data interface{}, signature, publicKey string) (bool, error) {
	canonical, err := canonicalgojson.MarshalCanonicalGoJSON(data)
	if err != nil {
		return false, fmt.Errorf("Error marshalling data canonically: %v", err)
	}
	return VerifyString(string(canonical), signature, publicKey)
}
