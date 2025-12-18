FROM golang=1.25-alpine AS builder
RUN update-ca-certificates
WORKDIR /myapp
ENV GOCACHE=$HOME/.cache/go-build
RUN --mount=type=cache,target=$GOCACHE
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
-ldflags='-w -s -extldflags "-static"' -a \
-o /serve .
FROM gcr.io/distroless/static
COPY --from-builder /serve /serve
ENTRYPOINT ["serve"]
