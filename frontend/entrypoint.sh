#!/bin/sh
set -e

# Substitute PUBLIC_API_URL in built JS files if set
if [ -n "$PUBLIC_API_URL" ]; then
    find /usr/share/nginx/html -name '*.js' -exec \
        sed -i "s|__PUBLIC_API_URL__|$PUBLIC_API_URL|g" {} +
fi

exec nginx -g "daemon off;"
