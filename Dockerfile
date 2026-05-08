FROM golang:1.22-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /sbomforge ./cmd/

FROM alpine:3.21

RUN apk add --no-cache curl ca-certificates

RUN curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh \
    | sh -s -- -b /usr/local/bin

RUN curl -sSfL \
    https://github.com/sigstore/cosign/releases/latest/download/cosign-linux-amd64 \
    -o /usr/local/bin/cosign \
    && chmod +x /usr/local/bin/cosign

COPY --from=builder /sbomforge /sbomforge

ENTRYPOINT [ "/sbomforge" ]
