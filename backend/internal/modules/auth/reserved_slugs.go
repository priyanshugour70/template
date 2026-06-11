package auth

import "strings"

// reservedSlugs is the list of tenant slugs that cannot be registered. They
// either collide with infra subdomains (api, www, cdn) or with future
// product subdomains (admin, console, status). Keep in sync with the
// frontend list at src/lib/tenant/reserved.ts.
var reservedSlugs = map[string]struct{}{
	"api": {}, "www": {}, "admin": {}, "app": {}, "auth": {}, "mail": {},
	"smtp": {}, "imap": {}, "cdn": {}, "static": {}, "assets": {}, "docs": {},
	"status": {}, "support": {}, "help": {}, "blog": {}, "dashboard": {},
	"console": {}, "dev": {}, "staging": {}, "prod": {}, "production": {},
	"test": {}, "root": {}, "ws": {}, "mx": {}, "ftp": {}, "pop": {},
	"webhook": {}, "webhooks": {},
}

// IsReservedSlug reports whether the given slug cannot be used as a tenant
// subdomain. Case- and whitespace-insensitive.
func IsReservedSlug(slug string) bool {
	_, ok := reservedSlugs[strings.ToLower(strings.TrimSpace(slug))]
	return ok
}
