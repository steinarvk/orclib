package identity

import "github.com/steinarvk/orclib/lib/orckeys"

type IdentityClaim struct {
	Timestamp           string                   `json:"timestamp"`
	EphemeralID         string                   `json:"ephemeral_id"`
	ShortEphemeralID    string                   `json:"short_ephemeral_id"`
	ProgramName         string                   `json:"program"`
	CanonicalHost       string                   `json:"canonical_host,omitempty"`
	VersionInfo         interface{}              `json:"version"`
	EphemeralPublicKeys *orckeys.PublicKeyPacket `json:"ephemeral_public_keys"`
	PublicKeys          *orckeys.PublicKeyPacket `json:"public_keys"`
}
