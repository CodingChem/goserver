# Source
FROM debian:stable-slim

# Copy source
COPY goserver /bin/goserver

# Set port
ENV PORT=8080

# Start server
CMD ["/bin/goserver"]
