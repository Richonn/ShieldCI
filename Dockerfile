FROM golang:1.25-alpine@sha256:450ce2460f20b2f581cf1ac4f36606f88817e63f907f489d9a3b1c14ba821979 AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /shieldci ./cmd/shieldci

FROM alpine:3.19@sha256:b58899f069c47216f6002a6850143dc6fae0d35eb8b0df9300bbe6327b9c2171

RUN apk --no-cache add ca-certificates

COPY --from=builder /shieldci /shieldci

ENTRYPOINT ["/shieldci"]
