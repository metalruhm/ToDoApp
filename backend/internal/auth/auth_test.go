package auth

import "testing"

func TestPassword(t *testing.T) {
	hash, err := HashPassword("password123")
	if err != nil {
		t.Fatal(err)
	}
	if hash == "password123" {
		t.Fatal("hash must not equal plain password")
	}
	if !CheckPassword(hash, "password123") {
		t.Error("correct password should match")
	}
	if CheckPassword(hash, "wrong-password") {
		t.Error("wrong password should not match")
	}
}

func TestToken(t *testing.T) {
	token, hash, err := NewToken()
	if err != nil {
		t.Fatal(err)
	}
	if token == "" || hash == "" || token == hash {
		t.Fatalf("unexpected token/hash: %q %q", token, hash)
	}
	if HashToken(token) != hash {
		t.Error("HashToken(token) should equal the returned hash")
	}
	other, _, _ := NewToken()
	if other == token {
		t.Error("tokens must be unique")
	}
}
