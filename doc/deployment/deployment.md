# Deployment

Bourbon applications are compiled into single binary executables, making deployment straightforward.

## Building for Production

Compile your application using `go build`:

```bash
go build -o myapp main.go
```

This generates an executable file named `myapp`.

## Environment Variables

In production, avoid committing sensitive information like database passwords or secret keys. Use environment variables instead.

```bash
export BOURBON_APP_ENV="production"
export BOURBON_APP_DEBUG="false"
export BOURBON_DATABASE_PASSWORD="securepassword"
export BOURBON_APP_SECRET_KEY="long-random-string"
```

You can also use a `.env` file if you prefer (ensure it's in `.gitignore`).

## Reverse Proxy (Nginx)

It's recommended to run your Bourbon application behind a reverse proxy like Nginx to handle SSL termination and static file serving efficiently.

### Example Nginx Configuration

```nginx
server {
    listen 80;
    server_name example.com;

    location /static/ {
        alias /var/www/myapp/static/;
    }

    location / {
        proxy_pass http://localhost:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Systemd Service

To keep your application running in the background and restart automatically on failure, create a systemd service file.

Create `/etc/systemd/system/myapp.service`:

```ini
[Unit]
Description=My Bourbon App
After=network.target

[Service]
User=www-data
Group=www-data
WorkingDirectory=/var/www/myapp
ExecStart=/var/www/myapp/myapp
Restart=always
EnvironmentFile=/var/www/myapp/.env

[Install]
WantedBy=multi-user.target
```

Enable and start the service:

```bash
sudo systemctl enable myapp
sudo systemctl start myapp
```

## Docker

You can also containerize your application using Docker.

### Dockerfile

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o myapp main.go

# Final stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/myapp .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/settings.toml .

EXPOSE 8000
CMD ["./myapp"]
```

Build and run:

```bash
docker build -t myapp .
docker run -p 8000:8000 myapp
```
