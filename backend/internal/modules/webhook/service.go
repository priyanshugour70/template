package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	apperr "github.com/your-org/your-service/internal/pkg/errors"
)

type Service struct {
	repo *Repository
	log  *zap.Logger
	http *http.Client
}

func NewService(repo *Repository, log *zap.Logger) *Service {
	return &Service{
		repo: repo,
		log:  log,
		http: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *Service) Create(ctx context.Context, tenantID, orgID uuid.UUID, in CreateInput) (*CreateOutput, error) {
	if tenantID == uuid.Nil || orgID == uuid.Nil {
		return nil, apperr.New(apperr.CodeValidation, "tenant + org context required", nil)
	}
	secret, err := generateSecret()
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "generate secret failed", err)
	}
	events, _ := json.Marshal(in.Events)
	if events == nil {
		events = []byte("[]")
	}
	headers, _ := json.Marshal(in.Headers)
	if headers == nil {
		headers = []byte("{}")
	}

	w := &Webhook{
		Name:        in.Name,
		URL:         in.URL,
		Events:      events,
		Description: in.Description,
		Headers:     headers,
		SecretHash:  hashSecret(secret),
		IsActive:    true,
	}
	w.TenantID = tenantID
	w.OrganizationID = &orgID

	if err := s.repo.Create(ctx, w); err != nil {
		return nil, apperr.New(apperr.CodeInternal, "create webhook failed", err)
	}
	return &CreateOutput{Webhook: *w, Secret: secret}, nil
}

func (s *Service) Update(ctx context.Context, orgID, id uuid.UUID, in UpdateInput) (*Webhook, error) {
	if _, err := s.repo.Get(ctx, orgID, id); err != nil {
		if IsNotFound(err) {
			return nil, apperr.New(apperr.CodeNotFound, "webhook not found", nil)
		}
		return nil, apperr.New(apperr.CodeInternal, "load webhook failed", err)
	}
	patch := map[string]any{}
	if in.Name != nil {
		patch["name"] = *in.Name
	}
	if in.URL != nil {
		patch["url"] = *in.URL
	}
	if in.Description != nil {
		patch["description"] = *in.Description
	}
	if in.IsActive != nil {
		patch["is_active"] = *in.IsActive
		if *in.IsActive {
			patch["disabled_at"] = nil
			patch["disabled_reason"] = ""
		}
	}
	if in.Events != nil {
		b, _ := json.Marshal(in.Events)
		patch["events"] = b
	}
	if in.Headers != nil {
		b, _ := json.Marshal(in.Headers)
		patch["headers"] = b
	}
	if len(patch) > 0 {
		if err := s.repo.Update(ctx, id, patch); err != nil {
			return nil, apperr.New(apperr.CodeInternal, "update webhook failed", err)
		}
	}
	w, _ := s.repo.Get(ctx, orgID, id)
	return w, nil
}

func (s *Service) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	if _, err := s.repo.Get(ctx, orgID, id); err != nil {
		if IsNotFound(err) {
			return apperr.New(apperr.CodeNotFound, "webhook not found", nil)
		}
		return apperr.New(apperr.CodeInternal, "load webhook failed", err)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return apperr.New(apperr.CodeInternal, "delete webhook failed", err)
	}
	return nil
}

func (s *Service) List(ctx context.Context, orgID uuid.UUID) ([]Webhook, error) {
	rows, err := s.repo.List(ctx, orgID)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "list webhooks failed", err)
	}
	return rows, nil
}

func (s *Service) ListDeliveries(ctx context.Context, orgID, id uuid.UUID, limit int) ([]Delivery, error) {
	if _, err := s.repo.Get(ctx, orgID, id); err != nil {
		if IsNotFound(err) {
			return nil, apperr.New(apperr.CodeNotFound, "webhook not found", nil)
		}
		return nil, apperr.New(apperr.CodeInternal, "load webhook failed", err)
	}
	return s.repo.ListDeliveries(ctx, id, limit)
}

// TestFire synchronously POSTs the test payload to the webhook URL with the
// HMAC signature header, records the delivery, and returns the result.
func (s *Service) TestFire(ctx context.Context, orgID, id uuid.UUID, in TestFireInput) (*TestFireOutput, error) {
	w, err := s.repo.Get(ctx, orgID, id)
	if err != nil {
		if IsNotFound(err) {
			return nil, apperr.New(apperr.CodeNotFound, "webhook not found", nil)
		}
		return nil, apperr.New(apperr.CodeInternal, "load webhook failed", err)
	}
	event := in.Event
	if event == "" {
		event = "test.ping"
	}
	payload := in.Payload
	if payload == nil {
		payload = map[string]interface{}{"hello": "world", "test": true}
	}
	envelope := map[string]interface{}{
		"event":     event,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"data":      payload,
	}
	body, _ := json.Marshal(envelope)

	d := &Delivery{
		WebhookID: w.ID,
		Event:     event,
		Payload:   body,
		Status:    "pending",
		Attempt:   1,
	}

	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.URL, bytes.NewReader(body))
	if err != nil {
		d.Status = "failed"
		d.ErrorMessage = "build request failed: " + err.Error()
		_ = s.repo.RecordDelivery(ctx, d)
		return &TestFireOutput{Delivery: *d}, nil
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "your-service-webhook/0.1")
	// Custom headers
	var headers map[string]string
	if len(w.Headers) > 0 {
		_ = json.Unmarshal([]byte(w.Headers), &headers)
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}
	// HMAC signature (callers verify with their secret + raw body)
	// Note: we only have the secret HASH, not the plaintext. For the test-fire
	// case we sign with a placeholder; in production the worker would have the
	// plaintext at write time. To make this useful for the demo, we still send
	// a deterministic header so callers can verify their integration shape.
	sig := signPlaceholder(body, w.SecretHash)
	req.Header.Set("X-Signature-256", "sha256="+sig)
	req.Header.Set("X-Event-Type", event)

	resp, err := s.http.Do(req)
	d.DurationMs = ptrInt(int(time.Since(start).Milliseconds()))
	if err != nil {
		d.Status = "failed"
		d.ErrorMessage = err.Error()
		_ = s.repo.RecordDelivery(ctx, d)
		_ = s.repo.UpdateDeliveryStats(ctx, w.ID, 0, false)
		return &TestFireOutput{Delivery: *d}, nil
	}
	defer resp.Body.Close()

	rb, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	d.ResponseStatus = ptrInt(resp.StatusCode)
	d.ResponseBody = string(rb)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		d.Status = "success"
		now := time.Now()
		d.DeliveredAt = &now
		_ = s.repo.UpdateDeliveryStats(ctx, w.ID, resp.StatusCode, true)
	} else {
		d.Status = "failed"
		d.ErrorMessage = fmt.Sprintf("HTTP %d", resp.StatusCode)
		_ = s.repo.UpdateDeliveryStats(ctx, w.ID, resp.StatusCode, false)
	}
	_ = s.repo.RecordDelivery(ctx, d)
	return &TestFireOutput{Delivery: *d}, nil
}

// HashSecret is exported so consumers (worker) can compute the matching hash.
func HashSecret(plaintext string) string { return hashSecret(plaintext) }

func hashSecret(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(sum[:])
}

func signPlaceholder(body []byte, secretHash string) string {
	mac := hmac.New(sha256.New, []byte(secretHash))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func generateSecret() (string, error) {
	var b [24]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return "whsec_" + hex.EncodeToString(b[:]), nil
}

func ptrInt(v int) *int { return &v }
