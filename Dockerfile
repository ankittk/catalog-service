FROM golang:1.24-alpine AS builder

RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    curl \
    bash \
    make \
    protobuf \
    protobuf-dev

# Install buf CLI
RUN curl -sSL \
    "https://github.com/bufbuild/buf/releases/latest/download/buf-$(uname -s)-$(uname -m)" \
    -o "/usr/local/bin/buf" && \
    chmod +x "/usr/local/bin/buf"

# Install required protoc plugins
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest && \
    go install github.com/envoyproxy/protoc-gen-validate@latest && \
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

# Add Go bin to PATH
ENV PATH="/go/bin:${PATH}"

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN cd proto && buf generate

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o catalog-service ./cmd/server/main.go

FROM alpine:latest

RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    wget

RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/catalog-service .

COPY --from=builder /app/data ./data

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8000 9000

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8000/health || exit 1

CMD ["./catalog-service"] 