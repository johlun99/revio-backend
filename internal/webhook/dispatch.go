package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// ReviewEvent is the payload sent for review.created webhooks.
type ReviewEvent struct {
	Event  string       `json:"event"`
	Review ReviewDetail `json:"review"`
}

// ReviewDetail is the review data included in the webhook payload.
type ReviewDetail struct {
	ID               string  `json:"id"`
	ProductID        string  `json:"product_id"`
	TenantID         string  `json:"tenant_id"`
	AuthorName       string  `json:"author_name"`
	Rating           int16   `json:"rating"`
	Title            *string `json:"title,omitempty"`
	Body             string  `json:"body"`
	Status           string  `json:"status"`
	VerifiedPurchase bool    `json:"verified_purchase"`
	CreatedAt        string  `json:"created_at"`
}

var httpClient = &http.Client{Timeout: 10 * time.Second}

// Dispatch fires a review.created webhook in a background goroutine.
// It is a no-op if webhookURL is empty.
func Dispatch(webhookURL, signingKey string, detail ReviewDetail) {
	if webhookURL == "" {
		return
	}
	go func() {
		payload, err := json.Marshal(ReviewEvent{Event: "review.created", Review: detail})
		if err != nil {
			return
		}

		mac := hmac.New(sha256.New, []byte(signingKey))
		mac.Write(payload)
		sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(payload))
		if err != nil {
			slog.Warn("webhook: failed to build request", "url", webhookURL, "error", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Revio-Signature", sig)
		req.Header.Set("X-Revio-Event", "review.created")
		req.Header.Set("User-Agent", "Revio-Webhook/1.0")

		resp, err := httpClient.Do(req)
		if err != nil {
			slog.Warn("webhook: delivery failed", "url", webhookURL, "error", err)
			return
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode >= 400 {
			slog.Warn("webhook: endpoint returned error", "url", webhookURL, "status", resp.StatusCode)
		}
	}()
}

// UUIDString formats a pgtype.UUID as a standard hyphenated UUID string.
func UUIDString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	b := u.Bytes
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
