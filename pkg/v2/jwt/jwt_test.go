package jwt

import (
	"reflect"
	"testing"
)

func TestNewJwtToken(t *testing.T) {
	secret := "test_secret"
	data := map[string]interface{}{
		"username": "test_user",
	}

	token, err := NewJwtToken(secret, data)
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}

	if token == "" {
		t.Errorf("Expected non-empty string, got empty string")
	}
}

func TestParseJwtToken(t *testing.T) {
	secret := "test_secret"
	data := map[string]interface{}{
		"username": "test_user",
	}

	token, _ := NewJwtToken(secret, data)
	extract := []string{"username"}

	values, err := ParseJwtToken(secret, token, extract)
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}

	expected := []string{"test_user"}
	if !reflect.DeepEqual(values, expected) {
		t.Errorf("Expected %v, got %v", expected, values)
	}
}

func TestParseJwtTokenInvalid(t *testing.T) {
	secret := "test_secret"
	invalidToken := "invalid_token"
	extract := []string{"username"}

	_, err := ParseJwtToken(secret, invalidToken, extract)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}
