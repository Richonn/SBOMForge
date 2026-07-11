FROM --platform=$BUILDPLATFORM golang:1.26-alpine@sha256:0648ddfa35769070197ba1cdf22a16dc452caf9315e66b91791308a543baf229 AS builder

ARG TARGETARCH

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -ldflags="-w -s" -o /sbomforge ./cmd/sbomforge

FROM alpine:3.21@sha256:48b0309ca019d89d40f670aa1bc06e426dc0931948452e8491e3d65087abc07d

ARG TARGETARCH

RUN apk add --no-cache curl ca-certificates

RUN curl -sSfL \
    "https://github.com/anchore/syft/releases/download/v1.46.0/syft_1.46.0_linux_${TARGETARCH}.tar.gz" \
    -o /tmp/syft.tar.gz \
    && tar -xzf /tmp/syft.tar.gz -C /usr/local/bin syft \
    && rm /tmp/syft.tar.gz

RUN curl -sSfL \
    "https://github.com/sigstore/cosign/releases/download/v3.1.1/cosign-linux-${TARGETARCH}" \
    -o /usr/local/bin/cosign \
    && chmod +x /usr/local/bin/cosign

COPY --from=builder /sbomforge /sbomforge

ENTRYPOINT [ "/sbomforge" ]
