#!/bin/sh

# Replace environment variables in nginx.conf
envsubst '${SHORTENER_PORT} ${REDIRECTOR_PORT}' < /usr/local/openresty/nginx/conf/nginx.conf.template > /usr/local/openresty/nginx/conf/nginx.conf

# Start OpenResty
exec /usr/local/openresty/bin/openresty -g 'daemon off;'
