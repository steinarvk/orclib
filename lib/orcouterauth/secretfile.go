package orcouterauth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type legacySecret struct {
	Secret string `json:"hashedsecret"`
}

func loadSecret(filename string) (*Secret, error) {
	var rv Secret
	f := func() error {
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("Read error: %v", err)
		}

		if err := json.Unmarshal(data, &rv); err != nil {
			return fmt.Errorf("JSON unmarshal error: %v", err)
		}

		if rv.Secret != "" {
			return nil
		}

		var legacySecret legacySecret

		if err := json.Unmarshal(data, &legacySecret); err != nil {
			return fmt.Errorf("JSON unmarshal error: %v", err)
		}

		if legacySecret.Secret != "" {
			rv.Secret = legacySecret.Secret
			return nil
		}

		return fmt.Errorf("Unrecognized format")
	}

	if err := f(); err != nil {
		return nil, fmt.Errorf("Error loading secret from %q: %v", filename, err)
	}

	if rv.Name == "" {
		rv.Name = filename
	}

	if rv.Secret == "" {
		return nil, fmt.Errorf("Internal error loading secret from %q", filename)
	}

	return &rv, nil
}

func loadSecrets(filenames []string) ([]*Secret, error) {
	var rv []*Secret
	for _, filename := range filenames {
		secret, err := loadSecret(filename)
		if err != nil {
			return nil, err
		}
		rv = append(rv, secret)
	}
	return rv, nil
}
