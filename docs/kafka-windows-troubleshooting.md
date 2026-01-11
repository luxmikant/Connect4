# Kafka Windows Compilation Issues - Troubleshooting Guide

## Issue Description

When building the analytics service on Windows, you may encounter linking errors with the `confluent-kafka-go` library:

```
undefined reference to `__imp__vsnprintf_s'
undefined reference to `_setjmp'
```

This is a **common issue** with the `confluent-kafka-go` library on Windows systems due to C library compatibility between the pre-compiled librdkafka and your system's GCC/MinGW installation.

## Root Cause

The `confluent-kafka-go` library includes a pre-compiled Windows version of librdkafka that was built with a different C runtime than your current MinGW/MSYS2 installation. The missing symbols (`__imp__vsnprintf_s`, `_setjmp`) are from Microsoft's Visual C++ runtime, but your Go toolchain is using GCC/MinGW.

## Solutions (Multiple Options)

### Solution 1: Use Pure Go Kafka Library (Recommended for Development)

Replace `confluent-kafka-go` with a pure Go implementation that doesn't require CGO:

#### Step 1: Update go.mod
```bash
go mod edit -droprequire github.com/confluentinc/confluent-kafka-go/v2
go get github.com/segmentio/kafka-go@latest
```

#### Step 2: Update Analytics Service
Create a new Kafka client implementation using the pure Go library:

```go
// internal/analytics/kafka_client.go
package analytics

import (
    "context"
    "encoding/json"
    "time"
    
    "github.com/segmentio/kafka-go"
    "connect4-multiplayer/internal/config"
    "connect4-multiplayer/pkg/models"
)

type KafkaClient struct {
    writer *kafka.Writer
}

func NewKafkaClient(cfg config.KafkaConfig) *KafkaClient {
    writer := &kafka.Writer{
        Addr:         kafka.TCP(cfg.BootstrapServers),
        Topic:        cfg.Topic,
        Balancer:     &kafka.LeastBytes{},
        BatchTimeout: 10 * time.Millisecond,
        Transport: &kafka.Transport{
            SASL: kafka.SASLMechanism(kafka.PlainSASLMechanism{
                Username: cfg.APIKey,
                Password: cfg.APISecret,
            }),
            TLS: &tls.Config{},
        },
    }
    
    return &KafkaClient{writer: writer}
}

func (k *KafkaClient) SendEvent(ctx context.Context, event *models.GameEvent) error {
    data, err := json.Marshal(event)
    if err != nil {
        return err
    }
    
    return k.writer.WriteMessages(ctx, kafka.Message{
        Key:   []byte(event.GameID),
        Value: data,
    })
}

func (k *KafkaClient) Close() error {
    return k.writer.Close()
}
```

### Solution 2: Install TDM-GCC (Alternative Compiler)

TDM-GCC is often more compatible with pre-compiled Windows libraries:

#### Step 1: Install TDM-GCC
1. Download TDM-GCC from: https://jmeubank.github.io/tdm-gcc/
2. Install the 64-bit version
3. Add TDM-GCC to your PATH before MSYS2/MinGW

#### Step 2: Set Environment Variables
```powershell
$env:CC = "gcc"
$env:CXX = "g++"
$env:CGO_ENABLED = "1"
```

#### Step 3: Rebuild
```bash
go clean -cache
go build cmd/analytics/main.go
```

### Solution 3: Use Docker for Analytics Service

Run the analytics service in a Docker container to avoid Windows compilation issues:

#### Step 1: Create Analytics Dockerfile
```dockerfile
# Dockerfile.analytics
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o analytics cmd/analytics/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/analytics .
CMD ["./analytics"]
```

#### Step 2: Update docker-compose.yml
```yaml
services:
  analytics:
    build:
      context: .
      dockerfile: Dockerfile.analytics
    environment:
      - DATABASE_URL=${DATABASE_URL}
      - KAFKA_BOOTSTRAP_SERVERS=${KAFKA_BOOTSTRAP_SERVERS}
      - KAFKA_API_KEY=${KAFKA_API_KEY}
      - KAFKA_API_SECRET=${KAFKA_API_SECRET}
    depends_on:
      - server
```

### Solution 4: Use WSL2 (Windows Subsystem for Linux)

Develop in WSL2 where CGO compilation works more reliably:

#### Step 1: Install WSL2
```powershell
wsl --install
```

#### Step 2: Install Go in WSL2
```bash
# In WSL2 terminal
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

#### Step 3: Clone and Build in WSL2
```bash
git clone <your-repo>
cd connect4-multiplayer
go build cmd/analytics/main.go
```

### Solution 5: Disable Analytics Temporarily

For immediate development, you can disable the analytics service:

#### Step 1: Create Build Tags
Add build tags to exclude analytics:

```go
//go:build !noanalytics
// +build !noanalytics

// cmd/analytics/main.go
package main
// ... existing code
```

#### Step 2: Build Without Analytics
```bash
go build -tags noanalytics cmd/server/main.go
```

## Recommended Approach

**For Development**: Use **Solution 1** (Pure Go Kafka library) as it:
- ✅ Eliminates CGO dependencies
- ✅ Works reliably on all platforms
- ✅ Easier to debug and maintain
- ✅ Better performance in many cases
- ✅ No compilation issues

**For Production**: Use **Solution 3** (Docker) as it:
- ✅ Ensures consistent environment
- ✅ Avoids platform-specific issues
- ✅ Easier deployment and scaling

## Implementation Priority

1. **Immediate**: Use Solution 1 to unblock development
2. **Short-term**: Set up Docker containers for consistent builds
3. **Long-term**: Consider WSL2 for better Linux compatibility

## Testing the Fix

After implementing any solution, test with:

```bash
# Test compilation
go build cmd/analytics/main.go

# Test with your Confluent Cloud credentials
go run cmd/analytics/main.go
```

## Additional Notes

- This issue is **Windows-specific** and won't occur on Linux/macOS
- The confluent-kafka-go library is actively maintained, but Windows CGO issues persist
- Pure Go alternatives like `segmentio/kafka-go` are often preferred for cross-platform development
- Consider this when choosing libraries for future microservices

## Related Issues

- [confluent-kafka-go Windows Issues](https://github.com/confluentinc/confluent-kafka-go/issues?q=is%3Aissue+windows)
- [Go CGO Windows Compilation](https://github.com/golang/go/wiki/WindowsBuild)

---

**Status**: This is a known limitation, not a bug in your code. The main server and database functionality work perfectly.