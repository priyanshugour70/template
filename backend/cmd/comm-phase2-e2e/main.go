// Command comm-phase2-e2e drives a running API through the Phase-2
// communication module: ticket-issued WS upgrade, channel subscription,
// REST → WS message broadcast, typing fan-out, presence, and inbound
// webhook. Run against a freshly-booted API on localhost:8080.
//
//   go run ./cmd/comm-phase2-e2e
//
// Defaults work for the seeded admin user (admin@acme.example / Admin@123).
// Override base URL with COMM_E2E_BASE_URL.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
)

type apiClient struct {
	baseURL  string
	httpC    *http.Client
	token    string
	tenantID uuid.UUID
	userID   uuid.UUID
}

func newClient(base string) *apiClient {
	return &apiClient{
		baseURL: strings.TrimRight(base, "/"),
		httpC:   &http.Client{Timeout: 10 * time.Second},
	}
}

// ── REST helpers ──────────────────────────────────────────────────────────

func (c *apiClient) post(path string, body any, into any) error {
	return c.req(http.MethodPost, path, body, into)
}

func (c *apiClient) get(path string, into any) error {
	return c.req(http.MethodGet, path, nil, into)
}

func (c *apiClient) req(method, path string, body any, into any) error {
	var rdr *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(b)
	}
	var req *http.Request
	var err error
	if rdr != nil {
		req, err = http.NewRequest(method, c.baseURL+path, rdr)
	} else {
		req, err = http.NewRequest(method, c.baseURL+path, nil)
	}
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.httpC.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(resp.Body)
		return fmt.Errorf("%s %s: %d %s", method, path, resp.StatusCode, buf.String())
	}
	if into != nil {
		// Unwrap the envelope: {success, data: ...}
		var env struct {
			Success bool            `json:"success"`
			Data    json.RawMessage `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
			return err
		}
		if !env.Success {
			return fmt.Errorf("%s %s: success=false", method, path)
		}
		if into != nil && len(env.Data) > 0 {
			return json.Unmarshal(env.Data, into)
		}
	}
	return nil
}

// ── Auth flow ─────────────────────────────────────────────────────────────

type discoverResp struct {
	Tenants []struct {
		ID   uuid.UUID `json:"id"`
		Slug string    `json:"slug"`
	} `json:"tenants"`
}

type loginResp struct {
	AccessToken string `json:"accessToken"`
	User        struct {
		ID uuid.UUID `json:"id"`
	} `json:"user"`
}

func (c *apiClient) login(email, password string) error {
	var disc discoverResp
	if err := c.post("/api/v1/auth/discover", map[string]string{"email": email}, &disc); err != nil {
		return fmt.Errorf("discover: %w", err)
	}
	if len(disc.Tenants) == 0 {
		return fmt.Errorf("no tenant for %s — is the API seeded?", email)
	}
	c.tenantID = disc.Tenants[0].ID
	var login loginResp
	body := map[string]any{
		"email":    email,
		"password": password,
		"tenantId": c.tenantID,
	}
	if err := c.post("/api/v1/auth/login", body, &login); err != nil {
		return fmt.Errorf("login: %w", err)
	}
	c.token = login.AccessToken
	c.userID = login.User.ID
	return nil
}

// ── Comm helpers ─────────────────────────────────────────────────────────

type channel struct {
	ID uuid.UUID `json:"id"`
}

func (c *apiClient) createChannel(slug, name string) (uuid.UUID, error) {
	var conv channel
	body := map[string]any{
		"slug": slug,
		"name": name,
	}
	if err := c.post("/api/v1/comm/conversations/channels", body, &conv); err != nil {
		return uuid.Nil, fmt.Errorf("createChannel: %w", err)
	}
	return conv.ID, nil
}

func (c *apiClient) sendMessage(convID uuid.UUID, body string) (uuid.UUID, error) {
	var msg struct {
		ID uuid.UUID `json:"id"`
	}
	if err := c.post(
		fmt.Sprintf("/api/v1/comm/conversations/%s/messages", convID),
		map[string]any{"body": body},
		&msg,
	); err != nil {
		return uuid.Nil, fmt.Errorf("sendMessage: %w", err)
	}
	return msg.ID, nil
}

func (c *apiClient) issueTicket() (string, error) {
	var out struct {
		Ticket string `json:"ticket"`
	}
	if err := c.post("/api/v1/comm/ws/ticket", nil, &out); err != nil {
		return "", fmt.Errorf("issueTicket: %w", err)
	}
	return out.Ticket, nil
}

func (c *apiClient) createInboundHook(convID uuid.UUID, name string) (string, uuid.UUID, error) {
	var out struct {
		Hook  struct{ ID uuid.UUID `json:"id"` } `json:"hook"`
		Token string                              `json:"token"`
	}
	if err := c.post(
		fmt.Sprintf("/api/v1/comm/conversations/%s/hooks", convID),
		map[string]any{"name": name, "displayName": "Sentry"},
		&out,
	); err != nil {
		return "", uuid.Nil, fmt.Errorf("createHook: %w", err)
	}
	return out.Token, out.Hook.ID, nil
}

// ── WS client ─────────────────────────────────────────────────────────────

type wsClient struct {
	conn   *websocket.Conn
	frames chan map[string]any
	label  string
}

func dialWS(ctx context.Context, baseURL, ticket, label string) (*wsClient, error) {
	u := strings.Replace(baseURL, "http://", "ws://", 1)
	u = strings.Replace(u, "https://", "wss://", 1)
	u += "/api/v1/comm/ws?ticket=" + ticket
	conn, _, err := websocket.Dial(ctx, u, &websocket.DialOptions{})
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", label, err)
	}
	wc := &wsClient{conn: conn, frames: make(chan map[string]any, 64), label: label}
	go wc.readPump()
	return wc, nil
}

func (w *wsClient) readPump() {
	defer close(w.frames)
	for {
		_, data, err := w.conn.Read(context.Background())
		if err != nil {
			return
		}
		var f map[string]any
		if err := json.Unmarshal(data, &f); err != nil {
			continue
		}
		w.frames <- f
	}
}

func (w *wsClient) send(frame map[string]any) error {
	b, _ := json.Marshal(frame)
	return w.conn.Write(context.Background(), websocket.MessageText, b)
}

func (w *wsClient) close() { _ = w.conn.Close(websocket.StatusNormalClosure, "test done") }

// awaitFrame waits for a frame matching `predicate` until the timeout. Drops
// non-matching frames; useful when the hello/presence chatter races real
// events.
func (w *wsClient) awaitFrame(predicate func(map[string]any) bool, timeout time.Duration) (map[string]any, error) {
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	for {
		select {
		case <-deadline.C:
			return nil, fmt.Errorf("[%s] timed out waiting for frame after %s", w.label, timeout)
		case f, ok := <-w.frames:
			if !ok {
				return nil, fmt.Errorf("[%s] socket closed", w.label)
			}
			if predicate(f) {
				return f, nil
			}
			// Otherwise drop and keep waiting.
		}
	}
}

// ── Assertions ────────────────────────────────────────────────────────────

func check(label string, err error) {
	if err != nil {
		log.Fatalf("FAIL: %s — %v", label, err)
	}
	fmt.Printf("✓ %s\n", label)
}

// ── Main ──────────────────────────────────────────────────────────────────

func main() {
	base := os.Getenv("COMM_E2E_BASE_URL")
	if base == "" {
		base = "http://localhost:8080"
	}
	email := envOr("COMM_E2E_EMAIL", "admin@acme.example")
	password := envOr("COMM_E2E_PASSWORD", "Admin@123")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	c := newClient(base)
	check("login as "+email, c.login(email, password))

	suffix := uuid.NewString()[:8]
	convID, err := c.createChannel("phase2-e2e-"+suffix, "Phase 2 E2E "+suffix)
	check("create channel", err)
	fmt.Printf("  conversation id: %s\n", convID)

	// Two WS tickets to simulate two browser tabs of the same user.
	ticket1, err := c.issueTicket()
	check("issue ticket 1", err)
	ticket2, err := c.issueTicket()
	check("issue ticket 2", err)

	tab1, err := dialWS(ctx, base, ticket1, "tab1")
	check("ws connect tab1", err)
	defer tab1.close()
	tab2, err := dialWS(ctx, base, ticket2, "tab2")
	check("ws connect tab2", err)
	defer tab2.close()

	// Expect hello frame on both tabs first.
	_, err = tab1.awaitFrame(typeIs("hello"), 5*time.Second)
	check("tab1 receives hello", err)
	_, err = tab2.awaitFrame(typeIs("hello"), 5*time.Second)
	check("tab2 receives hello", err)

	// Both tabs subscribe to the new channel.
	check("tab1 subscribe", tab1.send(map[string]any{"type": "subscribe", "conversationId": convID}))
	check("tab2 subscribe", tab2.send(map[string]any{"type": "subscribe", "conversationId": convID}))
	// Small grace period for hub subscribe to land.
	time.Sleep(150 * time.Millisecond)

	// REST send → both tabs receive message.created.
	msgID, err := c.sendMessage(convID, "hello from REST")
	check("REST send message", err)
	fmt.Printf("  message id: %s\n", msgID)

	_, err = tab1.awaitFrame(func(f map[string]any) bool {
		return f["type"] == "message.created" && f["conversationId"] == convID.String()
	}, 5*time.Second)
	check("tab1 receives message.created", err)
	_, err = tab2.awaitFrame(func(f map[string]any) bool {
		return f["type"] == "message.created" && f["conversationId"] == convID.String()
	}, 5*time.Second)
	check("tab2 receives message.created", err)

	// Typing: tab1 sends typing, tab2 receives it; tab1 should NOT see its own.
	check("tab1 typing", tab1.send(map[string]any{"type": "typing", "conversationId": convID}))
	_, err = tab2.awaitFrame(func(f map[string]any) bool {
		return f["type"] == "typing" && f["conversationId"] == convID.String()
	}, 5*time.Second)
	check("tab2 receives typing", err)

	// Inbound webhook: create hook → POST plaintext message → both tabs receive.
	hookToken, hookID, err := c.createInboundHook(convID, "phase2 e2e hook")
	check("create inbound hook", err)
	fmt.Printf("  hook id: %s\n", hookID)

	resp, err := http.Post(
		base+"/api/v1/comm/inbound/"+hookToken,
		"application/json",
		bytes.NewReader([]byte(`{"text":"hello from webhook","username":"CI Bot"}`)),
	)
	check("POST inbound webhook", err)
	if resp.StatusCode >= 400 {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(resp.Body)
		log.Fatalf("FAIL: inbound webhook returned %d %s", resp.StatusCode, buf.String())
	}
	_ = resp.Body.Close()

	_, err = tab1.awaitFrame(func(f map[string]any) bool {
		if f["type"] != "message.created" {
			return false
		}
		msg, _ := f["message"].(map[string]any)
		body, _ := msg["body"].(string)
		return strings.Contains(body, "hello from webhook")
	}, 5*time.Second)
	check("tab1 receives webhook message.created", err)
	_, err = tab2.awaitFrame(func(f map[string]any) bool {
		if f["type"] != "message.created" {
			return false
		}
		msg, _ := f["message"].(map[string]any)
		body, _ := msg["body"].(string)
		return strings.Contains(body, "hello from webhook")
	}, 5*time.Second)
	check("tab2 receives webhook message.created", err)

	fmt.Println("\nALL CHECKS PASSED — phase 2 end-to-end OK")
}

func typeIs(t string) func(map[string]any) bool {
	return func(f map[string]any) bool { return f["type"] == t }
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
