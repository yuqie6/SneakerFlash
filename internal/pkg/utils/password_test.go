package utils

import "testing"

func TestHashPasswordAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("secret-123")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword() returned empty hash")
	}
	if !CheckPassword("secret-123", hash) {
		t.Fatal("CheckPassword() = false, want true")
	}
	if CheckPassword("wrong-password", hash) {
		t.Fatal("CheckPassword() = true for wrong password")
	}
}
