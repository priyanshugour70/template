package middleware

import (
	"net/url"
	"strings"
)

func redactQueryForLog(raw string) string {
	if raw == "" {
		return ""
	}
	q, err := url.ParseQuery(raw)
	if err != nil {
		return "[query_parse_error]"
	}
	for k := range q {
		lk := strings.ToLower(strings.TrimSpace(k))
		if sensitiveQueryKey(lk) {
			q[k] = []string{"[REDACTED]"}
		}
	}
	return q.Encode()
}

func sensitiveQueryKey(lowerKey string) bool {
	switch lowerKey {
	case "password", "passwd", "token", "refresh_token", "access_token", "id_token",
		"client_secret", "api_key", "authorization", "code":
		return true
	default:
		return strings.HasSuffix(lowerKey, "_token") || strings.Contains(lowerKey, "password")
	}
}
