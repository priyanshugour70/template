#!/usr/bin/env bash
# Configures Nginx as a TLS-terminating reverse proxy in front of the API.
# Idempotent: re-running on the same domain is safe.
set -euo pipefail

DOMAIN="${DOMAIN:?DOMAIN is required}"
UPSTREAM_PORT="${UPSTREAM_PORT:-8080}"
CERTBOT_EMAIL="${CERTBOT_EMAIL:-admin@example.com}"
SITE_NAME="${SITE_NAME:-app-backend}"

if [ "$(id -u)" -eq 0 ]; then
  SUDO=""
else
  SUDO="sudo"
fi

for bin in nginx curl; do
  if ! command -v "$bin" >/dev/null 2>&1; then
    echo "$bin is required before configuring Nginx" >&2
    exit 1
  fi
done

write_http_config() {
  $SUDO tee "/etc/nginx/sites-available/$SITE_NAME" >/dev/null <<NGINX
server {
    listen 80;
    listen [::]:80;
    server_name $DOMAIN;

    client_max_body_size 50m;

    location / {
        proxy_pass http://127.0.0.1:$UPSTREAM_PORT;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_read_timeout 120s;
        proxy_send_timeout 120s;
    }
}
NGINX
}

write_https_config() {
  $SUDO tee "/etc/nginx/sites-available/$SITE_NAME" >/dev/null <<NGINX
server {
    listen 80;
    listen [::]:80;
    server_name $DOMAIN;
    return 301 https://\$host\$request_uri;
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name $DOMAIN;

    ssl_certificate /etc/letsencrypt/live/$DOMAIN/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/$DOMAIN/privkey.pem;

    client_max_body_size 50m;

    location / {
        proxy_pass http://127.0.0.1:$UPSTREAM_PORT;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto https;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_read_timeout 120s;
        proxy_send_timeout 120s;
    }
}
NGINX
}

$SUDO mkdir -p /etc/nginx/sites-available /etc/nginx/sites-enabled
write_http_config
$SUDO ln -sfn "/etc/nginx/sites-available/$SITE_NAME" "/etc/nginx/sites-enabled/$SITE_NAME"
$SUDO rm -f /etc/nginx/sites-enabled/default
$SUDO nginx -t
$SUDO systemctl enable --now nginx >/dev/null 2>&1 || $SUDO service nginx start >/dev/null 2>&1 || true
$SUDO systemctl reload nginx >/dev/null 2>&1 || $SUDO service nginx reload >/dev/null 2>&1 || true

if command -v certbot >/dev/null 2>&1; then
  if $SUDO test -f "/etc/letsencrypt/live/$DOMAIN/fullchain.pem"; then
    echo "TLS certificate already exists for $DOMAIN"
  elif ! $SUDO certbot certonly --nginx -d "$DOMAIN" --non-interactive --agree-tos --email "$CERTBOT_EMAIL"; then
    echo "Warning: certbot failed for $DOMAIN. Check DNS A record and security group ports 80/443." >&2
    exit 0
  fi

  write_https_config
  $SUDO nginx -t
  $SUDO systemctl reload nginx >/dev/null 2>&1 || $SUDO service nginx reload >/dev/null 2>&1 || true
fi
