package cryptopacket

import (
	"bytes"
	"fmt"
	"time"

	"encoding/json"

	"github.com/steinarvk/orclib/lib/orckeys"
	"github.com/steinarvk/orclib/lib/orctimestamp"
)

type PublicKeyRegistry interface {
	LookupPublicKeys(string) (*orckeys.PublicKeyPacket, error)
}

func Pack(payload interface{}, keys *orckeys.Keys, registry PublicKeyRegistry, recipientOwner string) (string, error) {
	if recipientOwner == "" {
		return "", fmt.Errorf("Missing recipient")
	}
	recipientPubKey, err := registry.LookupPublicKeys(recipientOwner)
	if err != nil {
		return "", err
	}
	return packEncryptedOrUnencrypted(payload, keys, recipientPubKey)
}

func Unpack(target interface{}, ciphertext string, keys *orckeys.Keys, registry PublicKeyRegistry) (*Packet, error) {
	if registry == nil {
		return nil, fmt.Errorf("Missing public key registry")
	}
	return unpackGeneric(target, ciphertext, keys, registry)
}

func UnpackUnencrypted(target interface{}, plaintext string, registry PublicKeyRegistry) (*Packet, error) {
	if registry == nil {
		return nil, fmt.Errorf("Missing public key registry")
	}
	return unpackGeneric(target, plaintext, nil, registry)
}

func UnpackWithoutVerification(target interface{}, ciphertext string, keys *orckeys.Keys) (*Packet, error) {
	if keys == nil {
		return nil, fmt.Errorf("Missing server keys")
	}
	return unpackGeneric(target, ciphertext, keys, nil)
}

func unpackGeneric(target interface{}, packetdata string, keysForDecryption *orckeys.Keys, registryForVerification PublicKeyRegistry) (*Packet, error) {
	var packet Packet
	if keysForDecryption == nil {
		if err := json.Unmarshal([]byte(packetdata), &packet); err != nil {
			return nil, fmt.Errorf("Unable to unmarshal packet: %v", err)
		}
	} else {
		if err := DecryptIntoJSON(keysForDecryption.Decrypt, packetdata, &packet); err != nil {
			return nil, fmt.Errorf("Unable to decrypt packet: %v", err)
		}
	}

	if registryForVerification != nil {
		senderPubKey, err := registryForVerification.LookupPublicKeys(packet.Contents.Sender)
		if err != nil {
			return nil, fmt.Errorf("Error looking up packet sender %q: %v", packet.Contents.Sender, err)
		}

		ok, err := VerifyJSON(packet.Contents, packet.Signature, senderPubKey.PublicSigningKey)
		if !ok || err != nil {
			return nil, fmt.Errorf("Invalid signature by %q: %v", packet.Contents.Sender, err)
		}
	}

	if target != nil {
		data, err := json.Marshal(packet.Contents.Payload)
		if err != nil {
			return nil, fmt.Errorf("Unable to marshal data: %v", err)
		}
		if err := json.Unmarshal(data, target); err != nil {
			return nil, fmt.Errorf("Unable to unmarshal data: %v", err)
		}
	}

	return &packet, nil
}

func PackUnencrypted(payload interface{}, keys *orckeys.Keys) (string, error) {
	return packEncryptedOrUnencrypted(payload, keys, nil)
}

func PackUnencryptedJSON(payload interface{}, keys *orckeys.Keys) (*Packet, error) {
	marshalled, err := packEncryptedOrUnencrypted(payload, keys, nil)
	if err != nil {
		return nil, err
	}
	var rv Packet
	if err := json.Unmarshal([]byte(marshalled), &rv); err != nil {
		return nil, err
	}
	return &rv, nil
}

func packEncryptedOrUnencrypted(payload interface{}, keys *orckeys.Keys, encryptToRecipient *orckeys.PublicKeyPacket) (string, error) {
	contents := Contents{
		Timestamp: orctimestamp.Format(time.Now()),
		Sender:    keys.Metadata.Owner,
		Payload:   payload,
	}
	if encryptToRecipient != nil {
		contents.Recipient = encryptToRecipient.Metadata.Owner
	}
	signature, err := SignJSON(keys.Signer, contents)
	if err != nil {
		return "", err
	}
	wrapped := Packet{
		Contents:  contents,
		Signature: signature,
	}
	if encryptToRecipient != nil {
		encrypter, err := encryptToRecipient.EncryptTo()
		if err != nil {
			return "", err
		}
		return EncryptJSON(encrypter, wrapped)
	}

	// Return unencrypted
	unencryptedBuf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(unencryptedBuf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(wrapped); err != nil {
		return "", fmt.Errorf("Error serializing JSON: %v", err)
	}
	return unencryptedBuf.String(), nil
}
