# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy dependency files first
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build both binaries statically
# CGO_ENABLED=0 ensures the binary is statically linked and can run in scratch
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/translate ./cmd/translate && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/apiserver ./cmd/apiserver

# Final stage
FROM gcr.io/distroless/static-debian12

WORKDIR /

# Copy binaries from builder
COPY --from=builder /app/bin/translate /translate
COPY --from=builder /app/bin/apiserver /apiserver

# Expose the default port for the API server
EXPOSE 8080

# By default, run the API server
# You can override this to run the CLI tool instead:
# docker run <image> /translate [flags]
ENTRYPOINT ["/apiserver"]
