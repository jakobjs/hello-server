# hello-server

A small Go HTTP server for testing and experimentation. It handles graceful shutdown on `SIGTERM`/`SIGINT` and exposes the following endpoints:

| Endpoint | Description |
| --- | --- |
| `GET /` | Welcome message with usage hints |
| `GET /hello/{name}` | Returns `Hello, {name}!` |
| `GET /slow?seconds=N` | Responds after a delay (default 5s, max 10s) |
| `GET /healthz` | Liveness probe — fails for 10 min every 30 min after startup |
| `GET /ready` | Readiness probe — returns ready 2 min after startup |

The periodic health failures and delayed readiness are intentional, making this useful for testing Kubernetes probe behavior and rollout strategies.

## DockerHub

This container is pushed to <https://hub.docker.com/r/jakobjs/hello-server>

## Run locally

```bash
go run .
```

## Run with Docker

```bash
docker build -t hello-server .
docker run -p 8080:8080 hello-server
```

## Test

```bash
go test ./...
```
