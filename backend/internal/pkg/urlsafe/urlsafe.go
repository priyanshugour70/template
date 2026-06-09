// Package urlsafe validates that a URL is safe to fetch from the server
// without exposing internal networks or cloud metadata services.
//
// Used by webhook create/update + the delivery worker as defence-in-depth
// against SSRF (Server-Side Request Forgery).
package urlsafe

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
)

// ErrUnsafeURL is returned when the URL is rejected by the safety policy.
var ErrUnsafeURL = errors.New("unsafe url")

// blockedHosts is the case-insensitive literal-hostname denylist. Cloud
// metadata endpoints and well-known internal aliases.
var blockedHosts = map[string]bool{
	"localhost":                true,
	"localhost.localdomain":    true,
	"metadata.google.internal": true,
	"metadata":                 true,
	"instance-data":            true,
}

// blockedCIDRs are the IP ranges we refuse to fetch.
// Sourced from RFC 1918 (private), RFC 6890 (special-purpose), and well-known
// cloud-metadata endpoints (169.254.169.254 is in 169.254.0.0/16).
var blockedCIDRs = func() []*net.IPNet {
	cidrs := []string{
		"0.0.0.0/8",     // unspecified
		"10.0.0.0/8",    // RFC 1918 private
		"100.64.0.0/10", // shared address space (CGN)
		"127.0.0.0/8",   // loopback
		"169.254.0.0/16", // link-local + cloud metadata
		"172.16.0.0/12", // RFC 1918 private
		"192.168.0.0/16", // RFC 1918 private
		"::1/128",       // IPv6 loopback
		"fc00::/7",      // IPv6 unique-local
		"fe80::/10",     // IPv6 link-local
	}
	out := make([]*net.IPNet, 0, len(cidrs))
	for _, c := range cidrs {
		_, n, err := net.ParseCIDR(c)
		if err == nil {
			out = append(out, n)
		}
	}
	return out
}()

// Validate parses raw, resolves its host, and returns nil only if the URL is
// http(s) AND the host is neither a blocked hostname nor any IP in a blocked
// CIDR. On local dev set WEBHOOK_ALLOW_PRIVATE=true to disable the check.
func Validate(raw string) error {
	if os.Getenv("WEBHOOK_ALLOW_PRIVATE") == "true" {
		return nil
	}
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return fmt.Errorf("%w: parse: %v", ErrUnsafeURL, err)
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https":
		// OK
	default:
		return fmt.Errorf("%w: scheme %q not allowed", ErrUnsafeURL, u.Scheme)
	}
	host := strings.ToLower(u.Hostname())
	if host == "" {
		return fmt.Errorf("%w: missing host", ErrUnsafeURL)
	}
	if blockedHosts[host] {
		return fmt.Errorf("%w: host %q is internal", ErrUnsafeURL, host)
	}
	// Resolve to IPs and check each against the CIDR list. If DNS resolution
	// fails we err on the side of safety and reject.
	ips, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("%w: dns lookup: %v", ErrUnsafeURL, err)
	}
	for _, ip := range ips {
		for _, n := range blockedCIDRs {
			if n.Contains(ip) {
				return fmt.Errorf("%w: host %q resolves to blocked range %s", ErrUnsafeURL, host, n)
			}
		}
	}
	return nil
}
