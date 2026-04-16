FROM golang:1.25.9-alpine@sha256:04d017a27c481185c169884328a5761d052910fdced8c3b8edd686474efdf59b AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /shieldci ./cmd/shieldci

FROM alpine:3.21@sha256:c3f8e73fdb79deaebaa2037150150191b9dcbfba68b4a46d70103204c53f4709

RUN apk upgrade --no-cache && apk add --no-cache ca-certificates

COPY --from=builder /shieldci /shieldci

ENTRYPOINT ["/shieldci"]
