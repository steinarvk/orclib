package hashedsecret

// This is meant to provide a tiny bit of extra security over a hashed
// secret that is sent in plaintext over Basic auth.
// It is convenient because a token like this can be produced with
// standard tools on a normal shell:
//    $(t=$(date +%s); echo $t$(htpasswd -bnB '' "$t:${CANONICAL_HOST}:${SHARED_SECRET}"))
// Note that there is no protection against token reuse, and of course
// if the shared secret leaks (highly likely if using it on the shell!)
// everything is moot.
// This is NOT meant to be relied on as a sole layer of security,
// just as an outer layer meant to provide some relief from a
// non-determined attacker.

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	DefaultAcceptPast   = 1 * time.Minute
	DefaultAcceptFuture = 1 * time.Minute
)

type Generator struct {
	SharedSecret        string
	TargetCanonicalHost string
}

func (g *Generator) Username() string {
	return ""
}

func formatPassword(t time.Time, canonicalHost string, sharedSecret string) string {
	return fmt.Sprintf("%d:%s:%s", t.Unix(), canonicalHost, sharedSecret)
}

func (g *Generator) Generate(now time.Time) (string, error) {
	if g.SharedSecret == "" {
		return "", errors.New("no secret set")
	}
	if g.TargetCanonicalHost == "" {
		return "", errors.New("no target canonical host set")
	}
	password := formatPassword(now, g.TargetCanonicalHost, g.SharedSecret)
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d:%s", now.Unix(), hashed), nil
}

func (g *Generator) IsSafeForUntrustedTarget() bool {
	return true
}

type Verifier struct {
	SharedSecret  string
	CanonicalHost string
	AcceptPast    time.Duration
	AcceptFuture  time.Duration
}

type VerificationError struct {
	UnderlyingError error
}

func (e VerificationError) Error() string { return e.UnderlyingError.Error() }
func (e VerificationError) HttpCode() int { return 401 }

var (
	tokenRE = regexp.MustCompile(`^([0-9]+):(.+)$`)
)

func (v *Verifier) earliestOK(now time.Time) time.Time {
	dur := v.AcceptPast
	if dur == 0 {
		dur = DefaultAcceptPast
	}
	return now.Add(-dur)
}

func (v *Verifier) latestOK(now time.Time) time.Time {
	dur := v.AcceptFuture
	if dur == 0 {
		dur = DefaultAcceptFuture
	}
	return now.Add(dur)
}

func (v *Verifier) verify(now time.Time, token string) error {
	if v.SharedSecret == "" {
		return errors.New("no secret set")
	}
	if v.CanonicalHost == "" {
		return errors.New("no canonical host set")
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return errors.New("no token")
	}
	groups := tokenRE.FindStringSubmatch(token)
	if groups == nil {
		return errors.New("malformed token (bad format)")
	}
	n, err := strconv.ParseInt(groups[1], 10, 64)
	if err != nil {
		return errors.New("malformed token (bad number)")
	}
	t := time.Unix(n, 0)
	hashedPassword := []byte(groups[2])
	password := formatPassword(t, v.CanonicalHost, v.SharedSecret)
	if err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(password)); err != nil {
		return fmt.Errorf("invalid token")
	}
	if v.earliestOK(now).After(t) {
		return errors.New("token expired")
	}
	if v.latestOK(now).Before(t) {
		return errors.New("token is from the future")
	}
	return nil
}

func (v *Verifier) Verify(ignoredUsername string, now time.Time, token string) (bool, error) {
	err := v.verify(now, token)
	if err != nil {
		return false, VerificationError{err}
	}
	return true, nil
}
