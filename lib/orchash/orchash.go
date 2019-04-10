package orchash

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/steinarvk/orclib/lib/canonicalgojson"
)

const (
	hashKind = "sha256"
)

func ComputeHash(data []byte) string {
	h := sha256.New()
	h.Write(data)
	return fmt.Sprintf("%s:%x", hashKind, h.Sum(nil))
}

func VerifyHash(h string, data []byte) (bool, error) {
	components := strings.Split(h, ":")
	if len(components) < 1 || components[0] != hashKind {
		return false, fmt.Errorf("expected %q hash", hashKind)
	}
	return ComputeHash(data) == h, nil
}

func MustComputeJSONHash(data interface{}) string {
	h, err := ComputeJSONHash(data)
	if err != nil {
		logrus.Fatalf("Failed to compute JSON hash: %v", err)
	}
	return h
}

func ComputeJSONHash(data interface{}) (string, error) {
	serialized, err := canonicalgojson.MarshalCanonicalGoJSON(data)
	if err != nil {
		return "", fmt.Errorf("Error serializing data: %v", err)
	}
	return ComputeHash(serialized), nil
}

func VerifyJSONHash(h string, data interface{}) (bool, error) {
	serialized, err := canonicalgojson.MarshalCanonicalGoJSON(data)
	if err != nil {
		return false, fmt.Errorf("Error serializing data: %v", err)
	}
	return VerifyHash(h, serialized)
}
