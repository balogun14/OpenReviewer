package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"testing"
)

func TestValidateSignature(t *testing.T) {
	body := []byte(`{"ok":true}`)
	req, err := http.NewRequest(http.MethodPost, "/webhooks/github", strings.NewReader(string(body)))
	if err != nil {
		t.Fatal(err)
	}

	mac := hmac.New(sha256.New, []byte("secret"))
	_, _ = mac.Write(body)
	req.Header.Set("X-Hub-Signature-256", "sha256="+hex.EncodeToString(mac.Sum(nil)))

	if err := ValidateSignature(req, body, "secret"); err != nil {
		t.Fatalf("expected valid signature: %v", err)
	}
}

func TestValidateSignatureRejectsInvalidSignature(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "/webhooks/github", strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-Hub-Signature-256", "sha256=deadbeef")

	if err := ValidateSignature(req, []byte("{}"), "secret"); err == nil {
		t.Fatal("expected invalid signature error")
	}
}
