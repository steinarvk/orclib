package identity

import "fmt"

func validateIdentity(claim *IdentityClaim) error {
	if claim == nil {
		return fmt.Errorf("Identity not set")
	}
	if claim.ProgramName == "" {
		return fmt.Errorf("ProgramName not set")
	}
	if claim.VersionInfo == nil {
		return fmt.Errorf("VersionInfo not set")
	}
	if claim.CanonicalHost == "" {
		return fmt.Errorf("CanonicalHost not set")
	}
	return nil
}
