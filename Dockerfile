FROM cgr.dev/chainguard/go AS build

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN go mod verify

RUN CGO_ENABLED=0 go build -o /go/bin/app ./cmd/connector

FROM cgr.dev/chainguard/glibc-dynamic

# Copy application binary
COPY --from=build --chown=nonroot:nonroot /go/bin/app /
USER nonroot

ENTRYPOINT ["/app"]
