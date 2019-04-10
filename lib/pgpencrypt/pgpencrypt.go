package pgpencrypt

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"golang.org/x/crypto/openpgp"
)

type Config struct {
	KeyringURLs      []string
	KeyringFilenames []string
	Recipients       []string
}

type Keyring struct {
	entities []*openpgp.Entity
}

type Encrypter struct {
	targetEntities []*openpgp.Entity
	recipients     []string
}

type Armoring bool

const (
	Armored = Armoring(true)
	Binary  = Armoring(false)
)

func FromConfig(cfg *Config) (*Encrypter, error) {
	if cfg == nil || len(cfg.Recipients) == 0 {
		return nil, nil
	}
	kr := &Keyring{}
	for _, filename := range cfg.KeyringFilenames {
		if err := kr.AddFromFile(filename); err != nil {
			return nil, err
		}
	}
	for _, url := range cfg.KeyringURLs {
		if err := kr.AddFromURL(url); err != nil {
			return nil, err
		}
	}
	return kr.EncrypterTo(cfg.Recipients)
}

func readKeysFromReaderAs(r io.Reader, armored Armoring) ([]*openpgp.Entity, error) {
	var entities openpgp.EntityList
	var err error
	if armored {
		entities, err = openpgp.ReadArmoredKeyRing(r)
	} else {
		entities, err = openpgp.ReadKeyRing(r)
	}
	if err != nil {
		return nil, err
	}
	return ([]*openpgp.Entity)(entities), nil
}

func readKeysFromURL(url string) ([]*openpgp.Entity, error) {
	if !strings.HasPrefix(url, "https://") {
		return nil, fmt.Errorf("Will not read keys from %q: HTTPS required", url)
	}
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Failed to read keys from %q: %v", url, err)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read keys from %q: %v", url, err)
	}
	buf := bytes.NewBuffer(data)
	rv, err := readKeysFromReaderAs(buf, Armored)
	if err != nil {
		rv, err = readKeysFromReaderAs(buf, Binary)
	}
	if err != nil {
		return nil, fmt.Errorf("Failed to read keys from %q: %v", url, err)
	}
	return rv, nil
}

func readKeysFromFile(filename string) ([]*openpgp.Entity, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	rv, err := readKeysFromReaderAs(f, Armored)
	if err != nil {
		rv, err = readKeysFromReaderAs(f, Binary)
	}
	if err != nil {
		return nil, fmt.Errorf("Error reading keys from %q: %v", filename, err)
	}
	return rv, nil
}

func (k *Keyring) AddFromURL(url string) error {
	entities, err := readKeysFromURL(url)
	if err != nil {
		return err
	}
	k.entities = append(k.entities, entities...)
	return nil
}

func (k *Keyring) AddFromFile(filename string) error {
	entities, err := readKeysFromFile(filename)
	if err != nil {
		return err
	}
	k.entities = append(k.entities, entities...)
	return nil
}

func isEntityRecipient(entity *openpgp.Entity, recipient string) (bool, error) {
	recipient = strings.TrimSpace(recipient)
	if recipient == "" {
		return false, fmt.Errorf("empty recipient")
	}
	for _, identity := range entity.Identities {
		if identity.Name == recipient || identity.UserId.Name == recipient || identity.UserId.Email == recipient {
			return true, nil
		}
	}
	return false, nil
}

func (k *Keyring) findRecipient(recipient string) (*openpgp.Entity, error) {
	var rv []*openpgp.Entity
	for _, sel := range k.entities {
		ok, err := isEntityRecipient(sel, recipient)
		if err != nil {
			return nil, err
		}
		if ok {
			rv = append(rv, sel)
		}
	}
	if len(rv) == 0 {
		return nil, fmt.Errorf("recipient %q not found", recipient)
	}
	if len(rv) > 1 {
		return nil, fmt.Errorf("recipient %q was ambiguous", recipient)
	}
	return rv[0], nil
}

func (k *Keyring) findRecipients(recipients []string) ([]*openpgp.Entity, error) {
	var entities []*openpgp.Entity
	for _, recipient := range recipients {
		rv, err := k.findRecipient(recipient)
		if err != nil {
			return nil, err
		}
		entities = append(entities, rv)
	}
	return entities, nil
}

func (k *Keyring) EncrypterTo(recipients []string) (*Encrypter, error) {
	recipientEntities, err := k.findRecipients(recipients)
	if err != nil {
		return nil, err
	}
	rv := &Encrypter{
		targetEntities: recipientEntities,
		recipients:     recipients,
	}
	// Dummy operation to fail early for invalid keys.
	if _, err := rv.Encrypt(nil); err != nil {
		return nil, err
	}
	return rv, nil
}

func (e *Encrypter) Encrypt(data []byte) ([]byte, error) {
	buf := bytes.Buffer{}
	hints := &openpgp.FileHints{
		IsBinary: true,
	}
	plaintextwriter, err := openpgp.Encrypt(&buf, e.targetEntities, nil, hints, nil)
	if err != nil {
		return nil, err
	}
	n, err := plaintextwriter.Write(data)
	if err != nil {
		return nil, err
	}
	if n < len(data) {
		return nil, fmt.Errorf("short write during encryption: %d bytes < %d", n, len(data))
	}
	if err := plaintextwriter.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
