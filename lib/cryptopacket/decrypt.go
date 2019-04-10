package cryptopacket

import (
	"encoding/base64"
	"encoding/json"

	"github.com/google/tink/go/tink"
)

func DecryptString(dec tink.HybridDecrypt, data string) (string, error) {
	ciphertextData, err := base64.RawStdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	plaintext, err := dec.Decrypt(ciphertextData, nil)
	if err != nil {
		return "", nil
	}
	return string(plaintext), nil
}

func DecryptIntoJSON(dec tink.HybridDecrypt, ciphertext string, target interface{}) error {
	plaintext, err := DecryptString(dec, ciphertext)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(plaintext), target)
}
