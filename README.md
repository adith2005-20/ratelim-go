# Ratelim

Flexible Redis-backed rate limiter written in Go, inspired by [uber-go/ratelimit](https://github.com/uber-go/ratelimit)

### Features
- Sets RPS, Slack, and other Policy from a single config file.
- Uses Redis as the storage medium, can also be configured in the config file.
- Exposes Ratelimiting via REST API, making Ratelim framework-agnostic. Use it in any and all of your services!
- Efficient, low-latency rate limiting using atomic operations.

### Planned features
- [ ] Prometheus metrics and dashboard

### 🚀 Getting Started
```bash
git clone https://github.com/yourusername/ratelim.git
cd ratelim
go run main.go
```

Edit the config.yaml file to customize RPS, add Redis URL, and other params.
