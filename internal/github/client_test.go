package github

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientInstallationTokenAndPullRequestFiles(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	var sawTokenRequest bool
	var sawFilesRequest bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/app/installations/99/access_tokens":
			sawTokenRequest = true
			if r.Method != http.MethodPost {
				t.Fatalf("unexpected token method: %s", r.Method)
			}
			if r.Header.Get("Authorization") == "" {
				t.Fatal("missing app authorization header")
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"token": "installation-token"})
		case "/repos/acme/api/pulls/7/files":
			sawFilesRequest = true
			if r.URL.Query().Get("per_page") != "100" {
				t.Fatalf("unexpected per_page: %s", r.URL.Query().Get("per_page"))
			}
			if r.Header.Get("Authorization") != "Bearer installation-token" {
				t.Fatalf("unexpected authorization header: %q", r.Header.Get("Authorization"))
			}
			_ = json.NewEncoder(w).Encode([]PullRequestFile{
				{Filename: "main.go", Status: "modified", Patch: "@@ -1 +1 @@\n+package main"},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.String())
		}
	}))
	defer server.Close()

	client := NewClient(ClientConfig{
		BaseURL:       server.URL,
		AppID:         "123",
		PrivateKeyPEM: pemBytes,
	})
	client.now = func() time.Time { return time.Unix(1000, 0) }

	token, err := client.InstallationToken(t.Context(), 99)
	if err != nil {
		t.Fatalf("InstallationToken returned error: %v", err)
	}

	files, err := client.PullRequestFiles(t.Context(), token, "acme", "api", 7)
	if err != nil {
		t.Fatalf("PullRequestFiles returned error: %v", err)
	}

	if token != "installation-token" {
		t.Fatalf("unexpected token: %q", token)
	}
	if len(files) != 1 || files[0].Filename != "main.go" {
		t.Fatalf("unexpected files: %+v", files)
	}
	if !sawTokenRequest || !sawFilesRequest {
		t.Fatalf("expected both token and files requests")
	}
}
