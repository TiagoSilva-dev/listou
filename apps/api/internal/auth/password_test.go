package auth

import "testing"

func TestPasswordRoundTrip(t *testing.T) {
	hash, err := HashPassword("correct horse")
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := VerifyPassword("correct horse", hash); err != nil || !ok {
		t.Fatalf("expected match, got %v %v", ok, err)
	}
	if ok, _ := VerifyPassword("wrong", hash); ok {
		t.Fatal("expected mismatch")
	}
	if _, err := VerifyPassword("x", "$bcrypt$nope"); err == nil {
		t.Fatal("expected malformed hash error")
	}
}

func TestRegisterValidate(t *testing.T) {
	in := RegisterInput{Name: "  Tiago ", Email: " Tiago@Example.COM ", Password: "12345678"}
	if err := in.Validate(); err != nil {
		t.Fatal(err)
	}
	if in.Email != "tiago@example.com" || in.Name != "Tiago" {
		t.Fatalf("not normalized: %+v", in)
	}
	bad := RegisterInput{Name: "T", Email: "nope", Password: "short"}
	if err := bad.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}
