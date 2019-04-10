package orckeys

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/google/tink/go/core/registry"
	"github.com/google/tink/go/hybrid"
	"github.com/google/tink/go/keyset"
	"github.com/google/tink/go/signature"
	"github.com/google/tink/go/tink"
	"github.com/steinarvk/orclib/lib/orctimestamp"
	"github.com/steinarvk/orclib/lib/uniqueid"
)

type Metadata struct {
	ID      string `json:"id"`
	Owner   string `json:"owner"`
	Created string `json:"creation_time"`
	Updated string `json:"update_time"`
}

type Key struct {
	PublicKey  string
	privateKey *keyset.Handle
}

type PublicKeyPacket struct {
	Metadata            Metadata `json:"metadata"`
	PublicSigningKey    string   `json:"public_signing_key"`
	PublicEncryptionKey string   `json:"public_encryption_key"`
}

type PrivateKeyPacket struct {
	Metadata             Metadata `json:"metadata"`
	Encrypted            bool     `json:"encrypted"`
	MasterKeyURI         string   `json:"master_key_uri"`
	PrivateSigningKey    []byte   `json:"private_signing_key"`
	PrivateEncryptionKey []byte   `json:"private_encryption_key"`
}

func (p PublicKeyPacket) EncryptTo() (tink.HybridEncrypt, error) {
	pubkey, err := stringToPublicKey(p.PublicEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("Failed to unpack public encryption key %q: %v", p.PublicEncryptionKey, err)
	}
	rv, err := hybrid.NewHybridEncrypt(pubkey)
	if err != nil {
		return nil, fmt.Errorf("Failed to create hybrid encryptor from public key: %v", err)
	}
	if _, err := rv.Encrypt(nil, nil); err != nil {
		return nil, fmt.Errorf("Failed smoke-test encryption with encryptor from public key: %v", err)
	}
	return rv, nil
}

func (p PublicKeyPacket) VerifyFrom() (tink.Verifier, error) {
	pubkey, err := stringToPublicKey(p.PublicSigningKey)
	if err != nil {
		return nil, fmt.Errorf("Failed to unpack public signing key %q: %v", p.PublicSigningKey, err)
	}
	return signature.NewVerifier(pubkey)
}

type Keys struct {
	Metadata      Metadata
	SigningKey    *Key
	EncryptionKey *Key
	Signer        tink.Signer
	Decrypt       tink.HybridDecrypt
}

func (k *Keys) Public() PublicKeyPacket {
	return PublicKeyPacket{
		Metadata:            k.Metadata,
		PublicSigningKey:    k.SigningKey.PublicKey,
		PublicEncryptionKey: k.EncryptionKey.PublicKey,
	}
}

func (k *Keys) sanityCheck() error {
	sig, err := k.Signer.Sign(nil)
	if err != nil {
		return fmt.Errorf("Failed to sign empty sequence: %v", err)
	}

	verifier, err := k.Public().VerifyFrom()
	if err != nil {
		return fmt.Errorf("Failed to make verifier for own public key: %v", err)
	}

	if err := verifier.Verify(sig, nil); err != nil {
		return fmt.Errorf("Failed to verify own signature: %v", err)
	}

	encrypter, err := k.Public().EncryptTo()
	if err != nil {
		return fmt.Errorf("Failed to make encrypter: %v", err)
	}

	ciphertext, err := encrypter.Encrypt(nil, nil)
	if err != nil {
		return fmt.Errorf("Failed to encrypt empty string: %v", err)
	}

	plaintext, err := k.Decrypt.Decrypt(ciphertext, nil)
	if err != nil {
		return fmt.Errorf("Failed to decrypt empty string: %v", err)
	}

	if len(plaintext) != 0 {
		return fmt.Errorf("Failed to decrypt empty string: resulted in %q", plaintext)
	}

	return nil
}

func fromKeyHandles(metadata Metadata, signingKey, encryptionKey *keyset.Handle) (*Keys, error) {
	publicSigningKey, err := publicKeyToString(signingKey)
	if err != nil {
		return nil, fmt.Errorf("Failed to format public signing key: %v", err)
	}

	publicEncryptionKey, err := publicKeyToString(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("Failed to format public encryption key: %v", err)
	}

	signer, err := signature.NewSigner(signingKey)
	if err != nil {
		return nil, fmt.Errorf("Failed to create signer object: %v", err)
	}

	decrypter, err := hybrid.NewHybridDecrypt(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("Failed to create decrypter object: %v", err)
	}

	rv := &Keys{
		Metadata: metadata,
		SigningKey: &Key{
			PublicKey:  publicSigningKey,
			privateKey: signingKey,
		},
		EncryptionKey: &Key{
			PublicKey:  publicEncryptionKey,
			privateKey: encryptionKey,
		},
		Signer:  signer,
		Decrypt: decrypter,
	}

	if err := rv.sanityCheck(); err != nil {
		return nil, fmt.Errorf("Server keys failed sanity check: %v", err)
	}

	return rv, nil
}

func Generate(owner string) (*Keys, error) {
	id, err := uniqueid.New()
	if err != nil {
		return nil, fmt.Errorf("Failed to generate unique ID: %v", err)
	}

	signingKey, err := keyset.NewHandle(DefaultSigningKeyTemplate)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate signing key: %v", err)
	}

	encryptionKey, err := keyset.NewHandle(DefaultEncryptionKeyTemplate)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate encryption key: %v", err)
	}

	t := time.Now()
	return fromKeyHandles(Metadata{
		ID:      id,
		Owner:   owner,
		Created: orctimestamp.Format(t),
		Updated: orctimestamp.Format(t),
	}, signingKey, encryptionKey)
}

func getMasterKey(masterKeyURI string) (tink.AEAD, error) {
	kmsClient, err := registry.GetKMSClient(masterKeyURI)
	if err != nil {
		return nil, fmt.Errorf("Failed to get KMS client for %q: %v", masterKeyURI, err)
	}
	masterKey, err := kmsClient.GetAEAD(masterKeyURI)
	if err != nil {
		return nil, fmt.Errorf("Failed to get KMS-based master key %q: %v", masterKeyURI, err)
	}
	return masterKey, nil
}

func (k *Keys) WriteEncrypted(w io.Writer, masterKeyURI string) error {
	masterKey, err := getMasterKey(masterKeyURI)
	if err != nil {
		return err
	}

	var signingKeyData, encryptionKeyData bytes.Buffer

	if err := k.SigningKey.privateKey.Write(keyset.NewBinaryWriter(&signingKeyData), masterKey); err != nil {
		return fmt.Errorf("Unable to write signing key (using master key %q): %v", masterKeyURI, err)
	}

	if err := k.EncryptionKey.privateKey.Write(keyset.NewBinaryWriter(&encryptionKeyData), masterKey); err != nil {
		return fmt.Errorf("Unable to write encryption key (using master key %q): %v", masterKeyURI, err)
	}

	packet := PrivateKeyPacket{
		Metadata:             k.Metadata,
		Encrypted:            true,
		MasterKeyURI:         masterKeyURI,
		PrivateSigningKey:    signingKeyData.Bytes(),
		PrivateEncryptionKey: encryptionKeyData.Bytes(),
	}

	return json.NewEncoder(w).Encode(packet)
}

func LoadEncrypted(r io.Reader, overrideMasterKeyURI string) (*Keys, error) {
	var packet PrivateKeyPacket
	if err := json.NewDecoder(r).Decode(&packet); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal JSON: %v", err)
	}

	if !packet.Encrypted {
		return nil, fmt.Errorf("Keys are not encrypted; LoadEncrypted() refusing to load")
	}

	masterKeyURI := packet.MasterKeyURI
	if overrideMasterKeyURI != "" {
		masterKeyURI = overrideMasterKeyURI
	}

	masterKey, err := getMasterKey(masterKeyURI)
	if err != nil {
		return nil, err
	}

	privateSigningKeyBuf := bytes.NewBuffer(packet.PrivateSigningKey)
	privateEncryptionKeyBuf := bytes.NewBuffer(packet.PrivateEncryptionKey)

	privateSigningKey, err := keyset.Read(keyset.NewBinaryReader(privateSigningKeyBuf), masterKey)
	if err != nil {
		return nil, fmt.Errorf("Unable to read private signing key (using master key %q): %v", masterKeyURI, err)
	}

	privateEncryptionKey, err := keyset.Read(keyset.NewBinaryReader(privateEncryptionKeyBuf), masterKey)
	if err != nil {
		return nil, fmt.Errorf("Unable to read private encryption key (using master key %q): %v", masterKeyURI, err)
	}

	return fromKeyHandles(packet.Metadata, privateSigningKey, privateEncryptionKey)
}
