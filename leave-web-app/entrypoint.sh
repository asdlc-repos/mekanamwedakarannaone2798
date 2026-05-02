#!/bin/sh
# Default URL if not injected by OpenChoreo
LEAVE_SERVICE_URL="${LEAVE_SERVICE_URL:-http://localhost:8080}"
# Strip trailing slash
LEAVE_SERVICE_URL="${LEAVE_SERVICE_URL%/}"
export LEAVE_SERVICE_URL

# Substitute env var into nginx config
envsubst '$LEAVE_SERVICE_URL' \
  < /etc/nginx/conf.d/default.conf.template \
  > /etc/nginx/conf.d/default.conf

# Write runtime env for SPA
cat > /usr/share/nginx/html/env.js <<EOF
window.RUNTIME_BACKEND_API_URL = "/api";
EOF

exec "$@"
