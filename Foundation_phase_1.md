# Phase 1 — Foundation (your homework)

This is the infrastructure work that has to happen **outside the codebase** before the multi-tenant subdomain system will work. Everything in this file is something you (or your DevOps person) does manually. The code in this repo has been updated to assume this foundation is in place.

Quick reference of what you need to provide:

| Thing | Production | Local dev |
|---|---|---|
| Apex domain | `lssgoo.com` | `lvh.me` (or `localhost`) |
| API host | `api.lssgoo.com` | `localhost:8080` |
| Tenant pattern | `<tenant>.lssgoo.com` | `<tenant>.lvh.me:3000` |
| Wildcard SSL cert | yes (`*.lssgoo.com` + `lssgoo.com`) | not needed |

---

## 1. DNS records

Set these in your DNS provider (Route53 / Cloudflare / etc.) on `lssgoo.com`:

```
A      lssgoo.com           <FRONTEND_IP>      # apex (marketing + login)
A      www.lssgoo.com       <FRONTEND_IP>      # canonical alias
A      api.lssgoo.com       <BACKEND_IP>       # Go backend (always this)
A      *.lssgoo.com         <FRONTEND_IP>      # wildcard — every tenant subdomain
```

If you use Cloudflare proxy: the wildcard `*.lssgoo.com` works on Pro plan and above. On Free plan, wildcards still resolve but proxy-orange-cloud is per-record.

If you use AWS Route53: use **A records with Alias** to ALB / CloudFront / Amplify. Wildcards work for free.

**Verify** with:
```bash
dig +short acme.lssgoo.com    # should return FRONTEND_IP
dig +short api.lssgoo.com     # should return BACKEND_IP
dig +short anything.lssgoo.com # should also return FRONTEND_IP (wildcard)
```

---

## 2. SSL certificates

You need **two** certs:

1. **Wildcard** for `*.lssgoo.com` — covers `acme.lssgoo.com`, `beta.lssgoo.com`, `api.lssgoo.com`, etc.
2. **Apex** for `lssgoo.com` — wildcards do NOT cover the apex.

Combine them into one SAN cert when issuing. Both Let's Encrypt and ACM support multi-SAN.

### Option A — Let's Encrypt (free, manual DNS challenge required for wildcards)

```bash
sudo certbot certonly --manual --preferred-challenges dns \
  -d lssgoo.com -d '*.lssgoo.com' \
  --email you@lssgoo.com --agree-tos
```

Renewals must be automated with DNS-01 — use a Certbot plugin matching your DNS provider (`certbot-dns-cloudflare`, `certbot-dns-route53`, etc.).

### Option B — AWS ACM (free, automatic, only works inside AWS)

In ACM console → Request public certificate → add both `lssgoo.com` and `*.lssgoo.com` → DNS validation. ACM auto-renews. Attach to your CloudFront distribution / ALB.

### Option C — Cloudflare (free, easiest if using Cloudflare DNS)

Cloudflare's Universal SSL automatically covers `lssgoo.com` and one wildcard level (`*.lssgoo.com`). Free, zero config. Just keep DNS records orange-clouded.

**Verify** with:
```bash
openssl s_client -connect acme.lssgoo.com:443 -servername acme.lssgoo.com < /dev/null \
  | openssl x509 -noout -subject -dates
```

---

## 3. Local dev — subdomains on localhost

You have two options. Pick one.

### Option A (recommended) — `lvh.me` (no setup at all)

`lvh.me` and `*.lvh.me` resolve to `127.0.0.1` automatically (it's a public DNS service). No /etc/hosts edits needed.

- Apex: `http://lvh.me:3000`
- Tenant: `http://acme.lvh.me:3000`
- API stays at: `http://localhost:8080`

**The codebase is configured to use `lvh.me` as the dev apex by default.**

If your network blocks DNS for `lvh.me` (some corp networks do), fall back to Option B.

### Option B — `/etc/hosts`

Add to `/etc/hosts` (one line per tenant you want to test):

```
127.0.0.1   lssgoo.test
127.0.0.1   acme.lssgoo.test
127.0.0.1   beta.lssgoo.test
127.0.0.1   palmonas.lssgoo.test
127.0.0.1   giva.lssgoo.test
```

Then set `NEXT_PUBLIC_APEX_DOMAIN=lssgoo.test` in `.env.local`.

You'll need to re-edit this file each time you create a new test tenant. That's why `lvh.me` is easier.

### Option C — `*.localhost` (works in modern browsers)

Chrome/Firefox/Safari resolve `*.localhost` to `127.0.0.1` automatically. Works without any setup.

- Apex: `http://localhost:3000`
- Tenant: `http://acme.localhost:3000`

**Caveat:** some tools (curl by default, some older libs) don't resolve `*.localhost`. `lvh.me` is more universally portable.

---

## 4. Reserved subdomains (do not let users register these as tenant slugs)

The codebase has this list hardcoded in two places (frontend + backend) — keep them in sync if you add more:

```
api, www, admin, app, auth, mail, smtp, imap,
cdn, static, assets, docs, status, support,
help, blog, dashboard, console, dev, staging,
prod, production, test, root, ws, mx
```

When a user tries to sign up with `slug=admin`, the backend rejects with `slug_reserved`.

---

## 5. CORS / cookie story (for your reference — code handles this)

- Backend allows any origin matching `*.lssgoo.com` and the apex.
- Frontend cookies have **no `Domain` attribute** — they default to the current host only, so `acme.lssgoo.com` cookies are invisible to `beta.lssgoo.com`. This is intentional: it lets the same browser hold sessions for multiple tenants at once.
- The apex (`lssgoo.com`) keeps only an ephemeral "discovery" cookie, never a session.

---

## 6. Production checklist (run before going live)

- [ ] DNS records resolve (`dig` returns expected IPs)
- [ ] SSL covers both `lssgoo.com` and `*.lssgoo.com` (browser shows green padlock on a random subdomain)
- [ ] `https://api.lssgoo.com/health/live` returns 200
- [ ] Backend `.env` has `CORS_ALLOWED_APEX=lssgoo.com` and `APP_BASE_DOMAIN=lssgoo.com`
- [ ] Backend `.env` has `AUTH_JWT_SECRET` set to a real ≥32-char secret (not the dev placeholder)
- [ ] Frontend `.env` has `NEXT_PUBLIC_APEX_DOMAIN=lssgoo.com` and `API_URL=https://api.lssgoo.com`
- [ ] Open `lssgoo.com/login` in a fresh browser → can complete signup → land on `<your-slug>.lssgoo.com/dashboard`
- [ ] Open `acme.lssgoo.com` and `beta.lssgoo.com` in two tabs → both stay signed in independently

---

## 7. What to do AFTER this foundation is set

Run the code changes in this repo (they're already in place):

1. **Backend** — set `CORS_ALLOWED_APEX=lssgoo.com` + `APP_BASE_DOMAIN=lssgoo.com` in `.env`, restart `make run-api`.
2. **Frontend** — set `NEXT_PUBLIC_APEX_DOMAIN=lssgoo.com` (or `lvh.me` for local) in `.env.local`, run `pnpm dev`.
3. Test the flow described in section 6.

If anything breaks, the codebase logs both the resolved subdomain and the tenant lookup result — check the proxy.ts logs and backend `INFO` lines.
