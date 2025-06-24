FROM openresty/openresty:1.21.4.1-alpine

# Install required packages
RUN apk add --no-cache gettext perl git wget curl

# Install additional Lua modules for rate limiting and health checks
RUN /usr/local/openresty/bin/opm install openresty/lua-resty-limit-traffic \
    && /usr/local/openresty/bin/opm install openresty/lua-resty-upstream-healthcheck

# Copy nginx configuration template and entrypoint script
COPY nginx.conf.template /usr/local/openresty/nginx/conf/nginx.conf.template
COPY entrypoint.sh /entrypoint.sh

# Make entrypoint script executable
RUN chmod +x /entrypoint.sh

# Create log directory
RUN mkdir -p /var/log/nginx

# Expose ports
EXPOSE 80 443

# Use entrypoint script
ENTRYPOINT ["/entrypoint.sh"]
