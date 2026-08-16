# syntax=docker/dockerfile:1
# This image is for Framework/Asus local deployment through Docker Compose or
# Portainer, not the managed web-preview runtime.
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY telemetry ./telemetry
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
    -trimpath \
    -ldflags='-s -w -buildid=' \
    -o /out/nodescope-server ./cmd/nodescope-server \
    && CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
    -trimpath \
    -ldflags='-s -w -buildid=' \
    -o /out/nodescope-probe ./cmd/nodescope-probe

FROM gcr.io/distroless/static-debian12:nonroot
LABEL org.opencontainers.image.licenses="Apache-2.0"
WORKDIR /app
COPY --from=builder /out/nodescope-server /app/nodescope-server
COPY --from=builder /out/nodescope-probe /app/nodescope-probe
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app/nodescope-server"]
