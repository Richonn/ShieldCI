FROM golang:1.25.9-alpine@sha256:04d017a27c481185c169884328a5761d052910fdced8c3b8edd686474efdf59b AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /shieldci ./cmd/shieldci

FROM alpine:3.19@sha256:6baf43584bcb78f2e5847d1de515f23499913ac9f12bdf834811a3145eb11ca1

RUN apk --no-cache add ca-certificates

COPY --from=builder /shieldci /shieldci

ENTRYPOINT ["/shieldci"]
