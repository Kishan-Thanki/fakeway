
# syntax=docker/dockerfile:1

ARG VERSION=v1.2.0

FROM --platform=$BUILDPLATFORM golang:1.27.3-alpine3.24@sha256:d200d2dd76b03a99b52651fd1325aff7dcd35422c35c01bac56571b0507f3deb AS builder

WORKDIR /src

COPY . .

ARG TARGETOS
ARG TARGETARCH
ARG VERSION

RUN CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    go build \
    -trimpath \
    -buildvcs=false \
    -ldflags="-s -w -X github.com/kishan-thanki/fakeway/internal/httpapi.Version=${VERSION}" \
    -o /fakeway . \
    && printf 'fakeway:x:65532:65532:Fakeway service user:/nonexistent:/sbin/nologin\n' > /tmp/passwd \
    && printf 'fakeway:x:65532:\n' > /tmp/group

FROM scratch

ARG VERSION

LABEL org.opencontainers.image.title="Fakeway" \
      org.opencontainers.image.description="Lightweight HTTP mock server" \
      org.opencontainers.image.source="https://github.com/Kishan-Thanki/fakeway" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.version="${VERSION}"

COPY --from=builder /fakeway /fakeway
COPY --from=builder /tmp/passwd /etc/passwd
COPY --from=builder /tmp/group /etc/group

USER fakeway:fakeway

EXPOSE 8080

ENTRYPOINT ["/fakeway"]
