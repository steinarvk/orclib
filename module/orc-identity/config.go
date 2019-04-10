package identity

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/steinarvk/orclib/lib/orctimestamp"
	"github.com/steinarvk/orclib/lib/uniqueid"
)

type Config struct {
	ProgramName string
	VersionInfo interface{}
}

func (c Config) create() (*IdentityClaim, error) {
	if c.ProgramName == "" {
		return nil, fmt.Errorf("No ProgramName provided")
	}
	newEphemeralID, err := uniqueid.New()
	if err != nil {
		return nil, fmt.Errorf("Unable to generate EphemeralID: %v", err)
	}

	hexHashOfEphemeralID := fmt.Sprintf("%x", sha256.Sum256([]byte(newEphemeralID)))
	shortEphemeralID := hexHashOfEphemeralID[:8]

	t := time.Now()
	return &IdentityClaim{
		Timestamp:        orctimestamp.Format(t),
		EphemeralID:      newEphemeralID,
		ShortEphemeralID: shortEphemeralID,
		ProgramName:      c.ProgramName,
		VersionInfo:      c.VersionInfo,
	}, nil
}
