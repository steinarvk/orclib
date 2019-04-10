package cryptopacket

import (
	"encoding/base64"
	"fmt"

	"github.com/google/tink/go/tink"
	"github.com/steinarvk/orclib/lib/canonicalgojson"
)

func EncryptString(enc tink.HybridEncrypt, data string) (string, error) {
	ciphertext, err := enc.Encrypt([]byte(data), nil)
	if err != nil {
		return "", nil
	}
	return base64.RawStdEncoding.EncodeToString(ciphertext), nil
}

func EncryptJSON(enc tink.HybridEncrypt, data interface{}) (string, error) {
	canonical, err := canonicalgojson.MarshalCanonicalGoJSON(data)
	if err != nil {
		return "", fmt.Errorf("Error marshalling data canonically: %v", err)
	}
	return EncryptString(enc, string(canonical))
}
