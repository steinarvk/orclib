package orckeys

import (
	"bytes"
	"encoding/base64"
	"fmt"

	"github.com/google/tink/go/keyset"
)

func publicKeyToString(handle *keyset.Handle) (string, error) {
	publicHandle, err := handle.Public()
	if err != nil {
		return "", fmt.Errorf("handle.Public() failed: %v", err)
	}
	publicKeyBuf := &bytes.Buffer{}
	if err := publicHandle.WriteWithNoSecrets(keyset.NewBinaryWriter(publicKeyBuf)); err != nil {
		return "", fmt.Errorf("WriteWithNoSecrets() failed: %v", err)
	}
	publicKey := base64.RawStdEncoding.EncodeToString(publicKeyBuf.Bytes())
	return publicKey, nil
}

func stringToPublicKey(packed string) (*keyset.Handle, error) {
	publicKeyData, err := base64.RawStdEncoding.DecodeString(packed)
	if err != nil {
		return nil, err
	}
	publicKeyBuf := bytes.NewBuffer(publicKeyData)
	handle, err := keyset.ReadWithNoSecrets(keyset.NewBinaryReader(publicKeyBuf))
	if err != nil {
		return nil, err
	}
	return handle, nil
}
