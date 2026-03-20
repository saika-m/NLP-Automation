# Production stage
FROM nginx:alpine

# Copy custom nginx config
COPY nginx.conf /etc/nginx/nginx.conf

# Copy static files directly
COPY index.html /usr/share/nginx/html/
COPY styles.css /usr/share/nginx/html/
COPY renderer.js /usr/share/nginx/html/
COPY main.js /usr/share/nginx/html/

# Create non-root user
RUN addgroup -g 1001 -S nlp-automation && \
    adduser -S -D -H -u 1001 -h /var/cache/nginx -s /sbin/nologin -G nlp-automation -g nlp-automation nlp-automation

# Change ownership of nginx directories
RUN chown -R nlp-automation:nlp-automation /var/cache/nginx && \
    chown -R nlp-automation:nlp-automation /var/log/nginx && \
    chown -R nlp-automation:nlp-automation /etc/nginx/conf.d

# Switch to non-root user
USER nlp-automation

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000 || exit 1

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
