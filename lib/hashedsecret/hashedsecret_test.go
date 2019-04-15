package hashedsecret

import (
	"testing"
	"time"
)

func TestHashedSecret(t *testing.T) {
	g := &Generator{
		SharedSecret:        "hunter2",
		TargetCanonicalHost: "foo.bar.example.com:1234",
	}
	v := &Verifier{
		SharedSecret:  g.SharedSecret,
		CanonicalHost: g.TargetCanonicalHost,
	}
	now := time.Unix(1234567890, 0)
	token, err := g.Generate(now)
	if err != nil {
		t.Fatal(err)
	}
	plusOne := now.Add(1 * time.Second)
	ok, err := v.Verify(plusOne, token)
	if !ok {
		t.Errorf("Verify(%v, %q) = false, %v want true, nil", plusOne, token, err)
	}
	plusMany := now.Add(7 * 24 * time.Hour)
	ok, err = v.Verify(plusMany, token)
	if ok || err.Error() != "token expired" {
		t.Errorf("Verify(%v, %q) = %v, %v want err: token expired", plusMany, token, ok, err)
	}
	minusMany := now.Add(-7 * 24 * time.Hour)
	ok, err = v.Verify(minusMany, token)
	if ok || err.Error() != "token is from the future" {
		t.Errorf("Verify(%v, %q) = %v, %v want err: token is from the future", minusMany, token, ok, err)
	}
}

func TestAgainstCommandLine(t *testing.T) {
	token := `1552517875:$2y$05$qL3A8ZqEhw/JiaKEAmvSCuYsjPL5JbzzrFai5YJG/WzE0/Y/ICoSe`
	v := &Verifier{
		SharedSecret:  "hunter2",
		CanonicalHost: "example.com:80",
	}
	now := time.Unix(1552517878, 0)
	ok, err := v.Verify(now, token)
	if !ok {
		t.Errorf("Verify(%v, %q) = false, %v want true, nil", now, token, err)
	}
}
