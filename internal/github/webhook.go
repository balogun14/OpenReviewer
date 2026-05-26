package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type PullRequestEvent struct {
	Action      string `json:"action"`
	Number      int    `json:"number"`
	PullRequest struct {
		Title string `json:"title"`
		Body  string `json:"body"`
		Head  struct {
			SHA string `json:"sha"`
		} `json:"head"`
	} `json:"pull_request"`
	Repository struct {
		Name  string `json:"name"`
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
		FullName string `json:"full_name"`
	} `json:"repository"`
	Installation struct {
		ID int64 `json:"id"`
	} `json:"installation"`
}

func DecodePullRequestEvent(body []byte) (PullRequestEvent, error) {
	var event PullRequestEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return PullRequestEvent{}, err
	}
	return event, nil
}

func IsPullRequestReviewAction(action string) bool {
	switch action {
	case "opened", "synchronize", "reopened", "ready_for_review":
		return true
	default:
		return false
	}
}

func ValidateSignature(r *http.Request, body []byte, secret string) error {
	if secret == "" {
		return nil
	}

	got := r.Header.Get("X-Hub-Signature-256")
	if got == "" {
		return errors.New("missing X-Hub-Signature-256 header")
	}

	const prefix = "sha256="
	if !strings.HasPrefix(got, prefix) {
		return errors.New("invalid signature format")
	}

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	want := make([]byte, hex.EncodedLen(mac.Size()))
	hex.Encode(want, mac.Sum(nil))

	gotBytes, err := hex.DecodeString(strings.TrimPrefix(got, prefix))
	if err != nil {
		return errors.New("invalid signature encoding")
	}

	wantBytes := mac.Sum(nil)
	if !hmac.Equal(gotBytes, wantBytes) {
		return errors.New("invalid webhook signature")
	}

	_ = want
	return nil
}
